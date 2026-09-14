from rl_runner.dqn import DQNConfig, train_dqn
from rl_runner.runner import RunRequest


def test_dqn_training_emits_real_bounded_episode_metrics() -> None:
    config = DQNConfig(
        episodes=4,
        hidden_units=8,
        gamma=0.95,
        learning_rate=0.01,
        replay_capacity=64,
        batch_size=4,
        target_sync_steps=8,
        epsilon_start=1.0,
        epsilon_end=0.1,
        epsilon_decay_episodes=3,
        moving_average_window=3,
    )
    request = RunRequest(
        run_id="70000000-0000-4000-8000-000000000005",
        environment_id="11111111-1111-4111-8111-111111111101",
        environment_version="1.0.0",
        policy_id="22222222-2222-4222-8222-222222222203",
        policy_sha256="0" * 64,
        seed=17,
        max_steps=24,
    )

    result = train_dqn(request, config)

    assert result.status == "succeeded"
    assert result.terminal_reason == "training_budget_completed"
    assert len(result.episodes) == 4
    assert [episode.episode_number for episode in result.episodes] == [1, 2, 3, 4]
    assert all(0 < episode.step_count <= 24 for episode in result.episodes)
    assert len(result.metrics) == 40
    assert {metric.metric for metric in result.metrics} >= {
        "episode_reward",
        "moving_average_reward",
        "epsilon",
        "loss",
        "success_rate",
    }
    epsilon = [metric.value for metric in result.metrics if metric.metric == "epsilon"]
    assert epsilon == sorted(epsilon, reverse=True)


def test_dqn_training_honors_cooperative_cancel() -> None:
    config = DQNConfig(
        episodes=2,
        hidden_units=8,
        gamma=0.95,
        learning_rate=0.01,
        replay_capacity=32,
        batch_size=4,
        target_sync_steps=8,
        epsilon_start=1.0,
        epsilon_end=0.1,
        epsilon_decay_episodes=2,
        moving_average_window=2,
    )
    request = RunRequest(
        run_id="70000000-0000-4000-8000-000000000006",
        environment_id="11111111-1111-4111-8111-111111111101",
        environment_version="1.0.0",
        policy_id="22222222-2222-4222-8222-222222222203",
        policy_sha256="0" * 64,
        seed=17,
        max_steps=24,
    )

    result = train_dqn(request, config, control=lambda: "cancel")

    assert result.status == "cancelled"
    assert result.episodes == ()
