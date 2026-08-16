from __future__ import annotations

import math
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "workers/rl-runner/src"))

from rl_runner.policies import EXPECTED_HASHES, WORLD_MODEL_ID, PolicyRegistry  # noqa: E402
from rl_runner.runner import RunRequest  # noqa: E402
from rl_runner.world_model import WorldModelConfig, run_world_model_rollout  # noqa: E402


def main() -> None:
    policy = PolicyRegistry(ROOT / "artifacts/policies").load(
        WORLD_MODEL_ID, EXPECTED_HASHES[WORLD_MODEL_ID]
    )
    if policy.world_model_config is None:
        raise SystemExit("world-model configuration was not loaded")
    request = RunRequest(
        run_id="70000000-0000-4000-8000-000000000007",
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
    transitions = result.episode.transitions
    errors = [transition.step_error for transition in transitions]
    accumulated = transitions[-1].accumulated_error if transitions else None
    if result.status != "succeeded" or result.terminal_reason != "rollout_budget_completed":
        raise SystemExit("world-model rollout did not complete its controlled budget")
    if len(transitions) != 11 or result.episode.terminal_reason != "goal_reached":
        raise SystemExit("world-model rollout did not retain the expected real trajectory")
    if not math.isclose(result.episode.total_reward, 8.64, abs_tol=1e-9):
        raise SystemExit("world-model real reward differs from the registered rollout")
    if errors != [0.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 0.0]:
        raise SystemExit("world-model per-step errors differ from the versioned profile")
    if accumulated != 9.0 or len(result.metrics) != 34:
        raise SystemExit("world-model accumulated evidence is incomplete")
    if not all(math.isfinite(metric.value) for metric in result.metrics):
        raise SystemExit("world-model emitted a non-finite metric")
    print("OK - real bounded world-model rollout: 11 steps | prediction vs actual | risk")
    print(
        "Evidence: 12 transition examples | first divergence step 2 | "
        "8.64 reward | 9.00 accumulated cell error | goal reached"
    )


if __name__ == "__main__":
    main()
