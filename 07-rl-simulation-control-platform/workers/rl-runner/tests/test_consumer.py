from __future__ import annotations

import json
from pathlib import Path

from rl_runner.consumer import RunnerConsumer
from rl_runner.messages import MessageEnvelope

PROJECT_ROOT = Path(__file__).resolve().parents[3]


class RecordingRedis:
    def __init__(self) -> None:
        self.values: dict[str, tuple[str, int]] = {}

    def set(self, key: str, value: str, *, ex: int) -> None:
        self.values[key] = (value, ex)


def command(message_type: str) -> MessageEnvelope:
    value = json.loads(
        (PROJECT_ROOT / "contracts" / "events" / "examples" / "run-requested.valid.json")
        .read_text(encoding="utf-8")
    )
    value["message_type"] = message_type
    return MessageEnvelope.from_mapping(value)


def test_durable_control_commands_refresh_the_cooperative_signal(tmp_path: Path) -> None:
    client = RecordingRedis()
    consumer = RunnerConsumer(client, tmp_path)  # type: ignore[arg-type]
    run_id = command("rl.run.pause-requested.v1").run_id

    consumer._apply_control(command("rl.run.pause-requested.v1"))
    assert client.values[f"rl.control:{run_id}"] == ("pause", 900)

    consumer._apply_control(command("rl.run.resume-requested.v1"))
    assert client.values[f"rl.control:{run_id}"] == ("run", 900)

    consumer._apply_control(command("rl.run.cancel-requested.v1"))
    assert client.values[f"rl.control:{run_id}"] == ("cancel", 900)
