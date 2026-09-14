from pathlib import Path

import pytest

from rl_runner.policies import EXPECTED_HASHES, WORLD_MODEL_ID, PolicyRegistry
from rl_runner.runner import RunRequest
from rl_runner.world_model import WorldModelConfig, run_world_model_rollout

PROJECT_ROOT = Path(__file__).resolve().parents[3]


def test_empirical_world_model_exposes_rollout_drift() -> None:
    policy = PolicyRegistry(PROJECT_ROOT / "artifacts" / "policies").load(
        WORLD_MODEL_ID, EXPECTED_HASHES[WORLD_MODEL_ID]
    )
    assert policy.world_model_config is not None
    request = RunRequest(
        run_id="70000000-0000-4000-8000-000000000008",
        environment_id=policy.environment_id,
        environment_version="1.0.0",
        policy_id=policy.policy_id,
        policy_sha256=policy.sha256,
        seed=23,
        max_steps=64,
    )

    result = run_world_model_rollout(
        request, WorldModelConfig.from_mapping(policy.world_model_config)
    )

    assert result.status == "succeeded"
    assert result.episode.step_count == 11
    assert result.episode.total_reward == pytest.approx(8.64)
    assert result.episode.collisions == 1
    assert result.episode.transitions[1].step_error == 1.0
    assert result.episode.transitions[-1].accumulated_error == 9.0
    assert result.episode.transitions[-1].next_state.as_dict() == {"row": 0, "column": 5}
    assert len(result.metrics) == 34


def test_world_model_requires_all_registered_actions() -> None:
    with pytest.raises(ValueError, match="cover every registered action"):
        WorldModelConfig.from_mapping(
            {
                "modelType": "empirical-action-delta",
                "trainingExamples": [
                    {
                        "state": {"row": 5, "column": 0},
                        "action": "right",
                        "nextState": {"row": 5, "column": 1},
                    }
                    for _ in range(4)
                ],
                "rolloutActions": ["right"],
            }
        )
