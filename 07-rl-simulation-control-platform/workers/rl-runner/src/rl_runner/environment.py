from __future__ import annotations

from dataclasses import dataclass
from typing import Final, Literal

Action = Literal["up", "right", "down", "left"]
ACTIONS: Final[tuple[Action, ...]] = ("up", "right", "down", "left")


@dataclass(frozen=True, slots=True)
class Coordinate:
    row: int
    column: int

    @property
    def key(self) -> str:
        return f"{self.row},{self.column}"

    def as_dict(self) -> dict[str, int]:
        return {"row": self.row, "column": self.column}


@dataclass(frozen=True, slots=True)
class StepResult:
    state: Coordinate
    action: Action
    next_state: Coordinate
    reward: float
    terminated: bool
    collision: bool


class GridworldEnvironment:
    """Gymnasium-compatible reset/step semantics without arbitrary environment loading."""

    rows: Final[int] = 6
    columns: Final[int] = 6
    observation_space: Final[str] = "Discrete(36)"
    action_space: Final[tuple[Action, ...]] = ACTIONS
    start: Final[Coordinate] = Coordinate(5, 0)
    goal: Final[Coordinate] = Coordinate(0, 5)
    obstacles: Final[frozenset[Coordinate]] = frozenset(
        {
            Coordinate(4, 1),
            Coordinate(3, 1),
            Coordinate(2, 3),
            Coordinate(1, 3),
            Coordinate(1, 4),
        }
    )
    step_reward: Final[float] = -0.04
    collision_reward: Final[float] = -1.0
    goal_reward: Final[float] = 10.0

    def __init__(self) -> None:
        self._position = self.start

    @property
    def position(self) -> Coordinate:
        return self._position

    def reset(self, *, seed: int | None = None) -> tuple[dict[str, int], dict[str, object]]:
        if seed is not None and seed < 0:
            raise ValueError("seed must be non-negative")
        self._position = self.start
        return self._position.as_dict(), {"seed": seed, "environment_version": "1.0.0"}

    def step(
        self, action: Action
    ) -> tuple[dict[str, int], float, bool, bool, dict[str, object]]:
        if action not in ACTIONS:
            raise ValueError("action is not registered")

        current = self._position
        offsets: dict[Action, tuple[int, int]] = {
            "up": (-1, 0),
            "right": (0, 1),
            "down": (1, 0),
            "left": (0, -1),
        }
        row_offset, column_offset = offsets[action]
        candidate = Coordinate(current.row + row_offset, current.column + column_offset)
        collision = not self._inside(candidate) or candidate in self.obstacles
        next_position = current if collision else candidate
        terminated = next_position == self.goal
        reward = (
            self.goal_reward
            if terminated
            else self.collision_reward
            if collision
            else self.step_reward
        )
        self._position = next_position
        return (
            next_position.as_dict(),
            reward,
            terminated,
            False,
            {"collision": collision, "state_key": next_position.key},
        )

    def _inside(self, coordinate: Coordinate) -> bool:
        return 0 <= coordinate.row < self.rows and 0 <= coordinate.column < self.columns
