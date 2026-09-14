from __future__ import annotations

from collections.abc import Callable
from dataclasses import dataclass
from typing import Any
from uuid import uuid4

from .environment import ACTIONS, Action, Coordinate, GridworldEnvironment
from .runner import ControlState, EpisodeResult, RunRequest, Transition, _utc_now


@dataclass(frozen=True, slots=True)
class TransitionExample:
    state: Coordinate
    action: Action
    next_state: Coordinate


@dataclass(frozen=True, slots=True)
class WorldModelConfig:
    model_type: str
    training_examples: tuple[TransitionExample, ...]
    rollout_actions: tuple[Action, ...]

    @classmethod
    def from_mapping(cls, value: dict[str, Any]) -> WorldModelConfig:
        examples: list[TransitionExample] = []
        for raw in value.get("trainingExamples", []):
            action = str(raw["action"])
            if action not in ACTIONS:
                raise ValueError("world-model example action is not registered")
            examples.append(
                TransitionExample(
                    state=_coordinate(raw["state"]),
                    action=action,
                    next_state=_coordinate(raw["nextState"]),
                )
            )
        actions = tuple(str(action) for action in value.get("rolloutActions", []))
        if any(action not in ACTIONS for action in actions):
            raise ValueError("world-model rollout action is not registered")
        config = cls(
            model_type=str(value.get("modelType", "")),
            training_examples=tuple(examples),
            rollout_actions=actions,
        )
        config.validate()
        return config

    def validate(self) -> None:
        if self.model_type != "empirical-action-delta":
            raise ValueError("world-model type is not allowlisted")
        if not 4 <= len(self.training_examples) <= 64:
            raise ValueError("world-model training dataset is outside the controlled boundary")
        if not 1 <= len(self.rollout_actions) <= 64:
            raise ValueError("world-model rollout budget is outside the controlled boundary")
        observed_actions = {example.action for example in self.training_examples}
        if observed_actions != set(ACTIONS):
            raise ValueError("world-model training dataset must cover every registered action")
        for example in self.training_examples:
            _validate_coordinate(example.state)
            _validate_coordinate(example.next_state)


@dataclass(frozen=True, slots=True)
class RolloutMetric:
    metric: str
    value: float
    unit: str
    step: int


@dataclass(frozen=True, slots=True)
class WorldModelRolloutResult:
    episode: EpisodeResult
    metrics: tuple[RolloutMetric, ...]
    status: str
    terminal_reason: str


class EmpiricalActionDeltaModel:
    def __init__(self, examples: tuple[TransitionExample, ...]) -> None:
        offsets: dict[Action, list[tuple[int, int]]] = {action: [] for action in ACTIONS}
        for example in examples:
            offsets[example.action].append(
                (
                    example.next_state.row - example.state.row,
                    example.next_state.column - example.state.column,
                )
            )
        self.offsets: dict[Action, tuple[int, int]] = {}
        for action, values in offsets.items():
            row_delta = sum(value[0] for value in values) / len(values)
            column_delta = sum(value[1] for value in values) / len(values)
            if not row_delta.is_integer() or not column_delta.is_integer():
                raise ValueError("world-model examples do not produce an integral action delta")
            self.offsets[action] = (int(row_delta), int(column_delta))

    def predict(self, state: Coordinate, action: Action) -> Coordinate:
        row_delta, column_delta = self.offsets[action]
        return Coordinate(
            row=min(max(state.row + row_delta, 0), GridworldEnvironment.rows - 1),
            column=min(max(state.column + column_delta, 0), GridworldEnvironment.columns - 1),
        )


def run_world_model_rollout(
    request: RunRequest,
    config: WorldModelConfig,
    *,
    control: Callable[[], ControlState] | None = None,
) -> WorldModelRolloutResult:
    config.validate()
    model = EmpiricalActionDeltaModel(config.training_examples)
    environment = GridworldEnvironment()
    environment.reset(seed=request.seed)
    predicted_state = environment.start
    transitions: list[Transition] = []
    metrics: list[RolloutMetric] = [
        RolloutMetric("training_examples", float(len(config.training_examples)), "count", 1)
    ]
    started_at = _utc_now()
    accumulated_error = 0.0
    total_reward = 0.0
    collisions = 0
    status = "truncated"
    terminal_reason = "max_steps"

    for step_index, action in enumerate(config.rollout_actions[: request.max_steps]):
        if control is not None and control() == "cancel":
            status = "cancelled"
            terminal_reason = "cancelled"
            break
        state = environment.position
        predicted_next_state = model.predict(predicted_state, action)
        observation, reward, terminated, _, info = environment.step(action)
        next_state = Coordinate(observation["row"], observation["column"])
        step_error = float(
            abs(predicted_next_state.row - next_state.row)
            + abs(predicted_next_state.column - next_state.column)
        )
        accumulated_error += step_error
        final_step = step_index + 1 == min(len(config.rollout_actions), request.max_steps)
        transitions.append(
            Transition(
                step_index=step_index,
                state=state,
                action=action,
                next_state=next_state,
                reward=reward,
                terminated=terminated,
                truncated=not terminated and final_step,
                sampled_at=_utc_now(),
                predicted_state=predicted_state,
                predicted_next_state=predicted_next_state,
                step_error=step_error,
                accumulated_error=accumulated_error,
                model_version="empirical-action-delta/1.0.0",
            )
        )
        step_number = step_index + 1
        metrics.extend(
            (
                RolloutMetric("prediction_error", step_error, "cells", step_number),
                RolloutMetric("accumulated_error", accumulated_error, "cells", step_number),
                RolloutMetric(
                    "rollout_risk", accumulated_error / step_number, "ratio", step_number
                ),
            )
        )
        total_reward += reward
        collisions += int(bool(info["collision"]))
        predicted_state = predicted_next_state
        if terminated:
            status = "succeeded"
            terminal_reason = "goal_reached"
            break

    episode = EpisodeResult(
        episode_id=str(uuid4()),
        episode_number=1,
        status=status,
        total_reward=total_reward,
        collisions=collisions,
        terminal_reason=terminal_reason,
        started_at=started_at,
        completed_at=_utc_now(),
        transitions=tuple(transitions),
    )
    return WorldModelRolloutResult(
        episode=episode,
        metrics=tuple(metrics),
        status="cancelled" if status == "cancelled" else "succeeded",
        terminal_reason="cancelled" if status == "cancelled" else "rollout_budget_completed",
    )


def _coordinate(value: dict[str, Any]) -> Coordinate:
    return Coordinate(row=int(value["row"]), column=int(value["column"]))


def _validate_coordinate(value: Coordinate) -> None:
    if not 0 <= value.row < GridworldEnvironment.rows:
        raise ValueError("world-model row is outside the registered environment")
    if not 0 <= value.column < GridworldEnvironment.columns:
        raise ValueError("world-model column is outside the registered environment")
