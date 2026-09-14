import json
from pathlib import Path

import pytest

from rl_runner.messages import MessageEnvelope, MessageValidationError

PROJECT_ROOT = Path(__file__).resolve().parents[3]


def test_valid_command_envelope_is_accepted() -> None:
    value = json.loads(
        (PROJECT_ROOT / "contracts" / "events" / "examples" / "run-requested.valid.json")
        .read_text(encoding="utf-8")
    )
    envelope = MessageEnvelope.from_mapping(value)
    assert envelope.message_type == "rl.run.requested.v1"
    assert envelope.schema_version == "1.0"


def test_arbitrary_message_type_is_rejected() -> None:
    value = json.loads(
        (PROJECT_ROOT / "contracts" / "events" / "examples" / "run-requested.valid.json")
        .read_text(encoding="utf-8")
    )
    value["message_type"] = "python.execute.arbitrary.v1"
    with pytest.raises(MessageValidationError, match="not allowlisted"):
        MessageEnvelope.from_mapping(value)
