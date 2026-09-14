from pathlib import Path

import pytest

from rl_runner.policies import EXPECTED_HASHES, Q_LEARNING_ID, PolicyRegistry
from rl_runner.runner import RunRequest, run_episode

PROJECT_ROOT = Path(__file__).resolve().parents[3]


def test_q_learning_episode_is_real_and_deterministic() -> None:
    registry = PolicyRegistry(PROJECT_ROOT / "artifacts" / "policies")
    policy = registry.load(Q_LEARNING_ID, EXPECTED_HASHES[Q_LEARNING_ID])
    request = RunRequest(
        run_id="70000000-0000-4000-8000-000000000004",
        environment_id="11111111-1111-4111-8111-111111111101",
        environment_version="1.0.0",
        policy_id=Q_LEARNING_ID,
        policy_sha256=policy.sha256,
        seed=37,
        max_steps=80,
    )

    result = run_episode(request, policy)

    assert result.status == "succeeded"
    assert result.terminal_reason == "goal_reached"
    assert result.step_count == 10
    assert result.collisions == 0
    assert result.total_reward == pytest.approx(9.64)
    assert result.transitions[-1].next_state.as_dict() == {"row": 0, "column": 5}


def test_registry_rejects_unregistered_policy_path() -> None:
    registry = PolicyRegistry(Path(PROJECT_ROOT / "artifacts" / "policies"))
    with pytest.raises(ValueError, match="not allowlisted"):
        registry.load("../../unsafe.pkl", "0" * 64)
