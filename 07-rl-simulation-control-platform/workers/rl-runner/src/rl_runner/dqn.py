# ruff: noqa: S311 - seeded pseudo-randomness is part of the experiment contract.

from __future__ import annotations

import copy
import random
from collections.abc import Callable
from dataclasses import dataclass
from typing import Any
from uuid import uuid4

from .environment import ACTIONS, Coordinate, GridworldEnvironment
from .runner import ControlState, EpisodeResult, RunRequest, Transition, _utc_now


@dataclass(frozen=True, slots=True)
class DQNConfig:
    episodes: int
    hidden_units: int
    gamma: float
    learning_rate: float
    replay_capacity: int
    batch_size: int
    target_sync_steps: int
    epsilon_start: float
    epsilon_end: float
    epsilon_decay_episodes: int
    moving_average_window: int

    @classmethod
    def from_mapping(cls, value: dict[str, Any]) -> DQNConfig:
        config = cls(
            episodes=int(value["episodes"]),
            hidden_units=int(value["hiddenUnits"]),
            gamma=float(value["gamma"]),
            learning_rate=float(value["learningRate"]),
            replay_capacity=int(value["replayCapacity"]),
            batch_size=int(value["batchSize"]),
            target_sync_steps=int(value["targetSyncSteps"]),
            epsilon_start=float(value["epsilonStart"]),
            epsilon_end=float(value["epsilonEnd"]),
            epsilon_decay_episodes=int(value["epsilonDecayEpisodes"]),
            moving_average_window=int(value["movingAverageWindow"]),
        )
        config.validate()
        return config

    def validate(self) -> None:
        if not 1 <= self.episodes <= 100:
            raise ValueError("DQN episodes must be between 1 and 100")
        if not 4 <= self.hidden_units <= 64:
            raise ValueError("DQN hidden units must be between 4 and 64")
        if not 0.0 <= self.gamma <= 1.0 or not 0.0 < self.learning_rate <= 0.1:
            raise ValueError("DQN gamma or learning rate is outside the controlled boundary")
        if not 1 <= self.batch_size <= self.replay_capacity <= 2048:
            raise ValueError("DQN replay configuration is invalid")
        if not 1 <= self.target_sync_steps <= 1000:
            raise ValueError("DQN target synchronization interval is invalid")
        if not 0.0 <= self.epsilon_end <= self.epsilon_start <= 1.0:
            raise ValueError("DQN epsilon schedule is invalid")
        if self.epsilon_decay_episodes < 1 or self.moving_average_window < 1:
            raise ValueError("DQN schedule windows must be positive")


@dataclass(frozen=True, slots=True)
class Experience:
    state_index: int
    action_index: int
    reward: float
    next_state_index: int
    terminated: bool


@dataclass(frozen=True, slots=True)
class TrainingMetric:
    metric: str
    value: float
    unit: str
    step: int


@dataclass(frozen=True, slots=True)
class DQNTrainingResult:
    episodes: tuple[EpisodeResult, ...]
    metrics: tuple[TrainingMetric, ...]
    status: str
    terminal_reason: str


class NeuralQNetwork:
    """Small one-hidden-layer Q-network for the controlled discrete environment."""

    def __init__(self, state_count: int, hidden_units: int, action_count: int, rng: random.Random):
        scale = 0.05
        self.w1 = [
            [rng.uniform(-scale, scale) for _ in range(hidden_units)]
            for _ in range(state_count)
        ]
        self.b1 = [0.0 for _ in range(hidden_units)]
        self.w2 = [
            [rng.uniform(-scale, scale) for _ in range(action_count)]
            for _ in range(hidden_units)
        ]
        self.b2 = [0.0 for _ in range(action_count)]

    def q_values(self, state_index: int) -> tuple[list[float], list[float]]:
        hidden = [
            max(0.0, weight + bias)
            for weight, bias in zip(self.w1[state_index], self.b1, strict=True)
        ]
        output = [
            self.b2[action]
            + sum(hidden[index] * self.w2[index][action] for index in range(len(hidden)))
            for action in range(len(self.b2))
        ]
        return output, hidden

    def train_batch(
        self,
        batch: list[Experience],
        target: NeuralQNetwork,
        *,
        gamma: float,
        learning_rate: float,
    ) -> float:
        total_loss = 0.0
        scale = 1.0 / len(batch)
        for sample in batch:
            predicted, hidden = self.q_values(sample.state_index)
            next_values, _ = target.q_values(sample.next_state_index)
            target_value = sample.reward
            if not sample.terminated:
                target_value += gamma * max(next_values)
            error = predicted[sample.action_index] - target_value
            total_loss += error * error
            output_gradient = 2.0 * error * scale
            hidden_gradients = [
                output_gradient * self.w2[index][sample.action_index]
                for index in range(len(hidden))
            ]
            for index, activation in enumerate(hidden):
                self.w2[index][sample.action_index] -= learning_rate * output_gradient * activation
            self.b2[sample.action_index] -= learning_rate * output_gradient
            for index, gradient in enumerate(hidden_gradients):
                if hidden[index] > 0.0:
                    self.w1[sample.state_index][index] -= learning_rate * gradient
                    self.b1[index] -= learning_rate * gradient
        return total_loss / len(batch)


def train_dqn(
    request: RunRequest,
    config: DQNConfig,
    *,
    control: Callable[[], ControlState] | None = None,
    on_episode: Callable[[EpisodeResult, tuple[TrainingMetric, ...]], None] | None = None,
) -> DQNTrainingResult:
    config.validate()
    rng = random.Random(request.seed)
    online = NeuralQNetwork(36, config.hidden_units, len(ACTIONS), rng)
    target = copy.deepcopy(online)
    replay: list[Experience] = []
    episode_results: list[EpisodeResult] = []
    metrics: list[TrainingMetric] = []
    rewards: list[float] = []
    optimizer_steps = 0
    successes = 0

    for episode_number in range(1, config.episodes + 1):
        control_state: ControlState = control() if control is not None else "run"
        if control_state == "cancel":
            return DQNTrainingResult(
                episodes=tuple(episode_results),
                metrics=tuple(metrics),
                status="cancelled",
                terminal_reason="cancelled",
            )
        epsilon = _epsilon_for_episode(config, episode_number)
        environment = GridworldEnvironment()
        environment.reset(seed=request.seed + episode_number)
        started_at = _utc_now()
        transitions: list[Transition] = []
        episode_losses: list[float] = []
        action_counts = [0 for _ in ACTIONS]
        total_reward = 0.0
        collisions = 0
        terminal_reason = "max_steps"
        status = "truncated"

        for step_index in range(request.max_steps):
            control_state = control() if control is not None else "run"
            if control_state == "cancel":
                status = "cancelled"
                terminal_reason = "cancelled"
                break
            state = environment.position
            state_index = _state_index(state)
            q_values, _ = online.q_values(state_index)
            if rng.random() < epsilon:
                action_index = rng.randrange(len(ACTIONS))
            else:
                action_index = max(range(len(q_values)), key=q_values.__getitem__)
            action = ACTIONS[action_index]
            observation, reward, terminated, _, info = environment.step(action)
            next_state = Coordinate(observation["row"], observation["column"])
            truncated = not terminated and step_index + 1 == request.max_steps
            transitions.append(
                Transition(
                    step_index=step_index,
                    state=state,
                    action=action,
                    next_state=next_state,
                    reward=reward,
                    terminated=terminated,
                    truncated=truncated,
                    sampled_at=_utc_now(),
                )
            )
            replay.append(
                Experience(state_index, action_index, reward, _state_index(next_state), terminated)
            )
            if len(replay) > config.replay_capacity:
                replay.pop(0)
            if len(replay) >= config.batch_size:
                batch = rng.sample(replay, config.batch_size)
                episode_losses.append(
                    online.train_batch(
                        batch,
                        target,
                        gamma=config.gamma,
                        learning_rate=config.learning_rate,
                    )
                )
                optimizer_steps += 1
                if optimizer_steps % config.target_sync_steps == 0:
                    target = copy.deepcopy(online)
            action_counts[action_index] += 1
            total_reward += reward
            collisions += int(bool(info["collision"]))
            if terminated:
                status = "succeeded"
                terminal_reason = "goal_reached"
                successes += 1
                break

        completed_at = _utc_now()
        result = EpisodeResult(
            episode_id=str(uuid4()),
            episode_number=episode_number,
            status=status,
            total_reward=total_reward,
            collisions=collisions,
            terminal_reason=terminal_reason,
            started_at=started_at,
            completed_at=completed_at,
            transitions=tuple(transitions),
        )
        episode_results.append(result)
        rewards.append(total_reward)
        window = rewards[-config.moving_average_window :]
        mean_loss = sum(episode_losses) / len(episode_losses) if episode_losses else 0.0
        values = (
            ("episode_reward", total_reward, "reward"),
            ("moving_average_reward", sum(window) / len(window), "reward"),
            ("epsilon", epsilon, "ratio"),
            ("loss", mean_loss, "mse"),
            ("episode_steps", result.step_count, "steps"),
            ("success_rate", successes / episode_number, "ratio"),
            ("action_up", action_counts[0], "count"),
            ("action_right", action_counts[1], "count"),
            ("action_down", action_counts[2], "count"),
            ("action_left", action_counts[3], "count"),
        )
        episode_metrics = tuple(
            TrainingMetric(metric=name, value=float(value), unit=unit, step=episode_number)
            for name, value, unit in values
        )
        metrics.extend(episode_metrics)
        if on_episode is not None:
            on_episode(result, episode_metrics)
        if status == "cancelled":
            return DQNTrainingResult(
                episodes=tuple(episode_results),
                metrics=tuple(metrics),
                status="cancelled",
                terminal_reason="cancelled",
            )

    return DQNTrainingResult(
        episodes=tuple(episode_results),
        metrics=tuple(metrics),
        status="succeeded",
        terminal_reason="training_budget_completed",
    )


def _epsilon_for_episode(config: DQNConfig, episode_number: int) -> float:
    progress = min((episode_number - 1) / config.epsilon_decay_episodes, 1.0)
    return config.epsilon_start + (config.epsilon_end - config.epsilon_start) * progress


def _state_index(coordinate: Coordinate) -> int:
    return coordinate.row * 6 + coordinate.column
