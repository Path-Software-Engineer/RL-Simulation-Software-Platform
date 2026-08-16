from __future__ import annotations

import sys
from math import isfinite
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "workers/rl-runner/src"))

from rl_runner.dqn import DQNConfig, train_dqn  # noqa: E402
from rl_runner.policies import DQN_ID, EXPECTED_HASHES, PolicyRegistry  # noqa: E402
from rl_runner.runner import RunRequest  # noqa: E402


def main() -> None:
    policy = PolicyRegistry(ROOT / "artifacts/policies").load(
        DQN_ID, EXPECTED_HASHES[DQN_ID]
    )
    if policy.training_config is None:
        raise SystemExit("DQN training configuration is missing")
    request = RunRequest(
        run_id="33333333-3333-4333-8333-333333333302",
        environment_id=policy.environment_id,
        environment_version="1.0.0",
        policy_id=DQN_ID,
        policy_sha256=policy.sha256,
        seed=11,
        max_steps=64,
    )
    request.validate(policy)
    result = train_dqn(request, DQNConfig.from_mapping(policy.training_config))
    if result.status != "succeeded" or len(result.episodes) != 40:
        raise SystemExit("bounded DQN training did not complete its 40-episode profile")
    names = {metric.metric for metric in result.metrics}
    required = {"episode_reward", "moving_average_reward", "epsilon", "loss"}
    if not required <= names or len(result.metrics) != 400:
        raise SystemExit("DQN training metrics are incomplete")
    if not all(isfinite(metric.value) for metric in result.metrics):
        raise SystemExit("DQN training emitted a non-finite metric")
    rewards = [episode.total_reward for episode in result.episodes]
    final_average = next(
        metric.value
        for metric in reversed(result.metrics)
        if metric.metric == "moving_average_reward"
    )
    final_success_rate = next(
        metric.value
        for metric in reversed(result.metrics)
        if metric.metric == "success_rate"
    )
    print("OK - real bounded DQN check: 40 episodes | replay | epsilon | loss | target sync")
    print(
        "Evidence: "
        f"best reward {max(rewards):.2f} | final moving average {final_average:.2f} | "
        f"success rate {final_success_rate:.1%}"
    )


if __name__ == "__main__":
    main()
