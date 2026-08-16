from rl_runner.environment import Coordinate, GridworldEnvironment


def test_reset_and_registered_path_reach_goal() -> None:
    environment = GridworldEnvironment()
    observation, info = environment.reset(seed=37)
    assert observation == {"row": 5, "column": 0}
    assert info["environment_version"] == "1.0.0"

    for action in ["up"] * 5 + ["right"] * 5:
        observation, reward, terminated, truncated, _ = environment.step(action)

    assert observation == {"row": 0, "column": 5}
    assert reward == 10.0
    assert terminated is True
    assert truncated is False


def test_collision_keeps_state_and_penalizes_action() -> None:
    environment = GridworldEnvironment()
    environment.reset(seed=1)
    observation, reward, terminated, _, info = environment.step("left")
    assert observation == Coordinate(5, 0).as_dict()
    assert reward == -1.0
    assert terminated is False
    assert info["collision"] is True
