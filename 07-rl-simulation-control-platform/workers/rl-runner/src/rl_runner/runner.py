from __future__ import annotations

from collections.abc import Callable
from dataclasses import dataclass
from datetime import UTC, datetime
from typing import Literal
from uuid import uuid4

from .environment import Coordinate, GridworldEnvironment
from .policies import PolicyArtifact

ControlState = Literal["run", "pause", "cancel"]


@dataclass(frozen=True, slots=True)
class RunRequest:
    run_id: str
    environment_id: str
    environment_version: str
    policy_id: str
    policy_sha256: str
    seed: int
    max_steps: int

    def validate(self, policy: PolicyArtifact) -> None:
        if self.environment_id != "11111111-1111-4111-8111-111111111101":
            raise ValueError("environment is not allowlisted")
        if self.environment_version != "1.0.0":
            raise ValueError("environment version is incompatible")
        if self.policy_id != policy.policy_id or self.policy_sha256 != policy.sha256:
            raise ValueError("policy identity differs from the verified artifact")
        if not 0 <= self.seed <= 2_147_483_647:
            raise ValueError("seed is outside the accepted range")
        if not 1 <= self.max_steps <= 200:
            raise ValueError("max_steps is outside the accepted range")


@dataclass(frozen=True, slots=True)
class Transition:
    step_index: int
    state: Coordinate
    action: str
    next_state: Coordinate
    reward: float
    terminated: bool
    truncated: bool
    sampled_at: str
    predicted_state: Coordinate | None = None
    predicted_next_state: Coordinate | None = None
    step_error: float | None = None
    accumulated_error: float | None = None
    model_version: str | None = None

    def as_dict(self) -> dict[str, object]:
        value: dict[str, object] = {
            "id": str(uuid4()),
            "step_index": self.step_index,
            "state": self.state.as_dict(),
            "action": self.action,
            "next_state": self.next_state.as_dict(),
            "reward": self.reward,
            "terminated": self.terminated,
            "truncated": self.truncated,
            "sampled_at": self.sampled_at,
        }
        if self.predicted_state is not None and self.predicted_next_state is not None:
            value.update(
                {
                    "predicted_state": self.predicted_state.as_dict(),
                    "predicted_next_state": self.predicted_next_state.as_dict(),
                    "step_error": self.step_error,
                    "accumulated_error": self.accumulated_error,
                    "model_version": self.model_version,
                }
            )
        return value


@dataclass(frozen=True, slots=True)
class EpisodeResult:
    episode_id: str
    episode_number: int
    status: str
    total_reward: float
    collisions: int
    terminal_reason: str
    started_at: str
    completed_at: str
    transitions: tuple[Transition, ...]

    @property
    def step_count(self) -> int:
        return len(self.transitions)

    def as_payload(self) -> dict[str, object]:
        return {
            "episode_id": self.episode_id,
            "episode_number": self.episode_number,
            "status": self.status,
            "total_reward": round(self.total_reward, 4),
            "step_count": self.step_count,
            "collisions": self.collisions,
            "terminal_reason": self.terminal_reason,
            "started_at": self.started_at,
            "completed_at": self.completed_at,
            "transitions": [transition.as_dict() for transition in self.transitions],
        }


def run_episode(
    request: RunRequest,
    policy: PolicyArtifact,
    *,
    control: Callable[[], ControlState] | None = None,
) -> EpisodeResult:
    request.validate(policy)
    environment = GridworldEnvironment()
    environment.reset(seed=request.seed)
    started_at = _utc_now()
    transitions: list[Transition] = []
    total_reward = 0.0
    collisions = 0
    terminal_reason = "max_steps"
    status = "truncated"

    for step_index in range(request.max_steps):
        if control is not None and control() == "cancel":
            terminal_reason = "cancelled"
            status = "cancelled"
            break
        state = environment.position
        action = policy.action_for(state)
        next_observation, reward, terminated, _, info = environment.step(action)
        truncated = not terminated and step_index + 1 == request.max_steps
        next_state = Coordinate(next_observation["row"], next_observation["column"])
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
        total_reward += reward
        collisions += int(bool(info["collision"]))
        if terminated:
            terminal_reason = "goal_reached"
            status = "succeeded"
            break

    return EpisodeResult(
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


def _utc_now() -> str:
    return datetime.now(UTC).isoformat().replace("+00:00", "Z")
