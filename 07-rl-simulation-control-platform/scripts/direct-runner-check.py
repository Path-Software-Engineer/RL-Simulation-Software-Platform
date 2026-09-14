from __future__ import annotations

import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "workers/rl-runner/src"))

from rl_runner.policies import PolicyRegistry  # noqa: E402
from rl_runner.runner import RunRequest, run_episode  # noqa: E402


def main() -> None:
    artifact_root = ROOT / "artifacts/policies"
    policy_id = "22222222-2222-4222-8222-222222222201"
    sha256 = "22d5faf9a94fcd05fdf31d2a1429a1b8f02d4f394f6cb61b95cc26351060927f"
    policy = PolicyRegistry(artifact_root).load(policy_id, sha256)
    result = run_episode(
        RunRequest(
            run_id="33333333-3333-4333-8333-333333333301",
            environment_id="11111111-1111-4111-8111-111111111101",
            environment_version="1.0.0",
            policy_id=policy_id,
            policy_sha256=sha256,
            seed=7,
            max_steps=64,
        ),
        policy,
    )
    expected = ("succeeded", "goal_reached", 10, 0, 9.64)
    observed = (
        result.status,
        result.terminal_reason,
        result.step_count,
        result.collisions,
        round(result.total_reward, 2),
    )
    if observed != expected:
        raise SystemExit(f"runner evidence differs: {observed!r} != {expected!r}")
    print("OK - real Gridworld episode: 10 steps | 0 collisions | 9.64 reward | goal reached")


if __name__ == "__main__":
    main()
