from __future__ import annotations

from dataclasses import dataclass
from datetime import UTC, datetime
from typing import Any, Final
from uuid import UUID, uuid4

ALLOWED_MESSAGE_TYPES: Final[frozenset[str]] = frozenset(
    {
        "rl.run.requested.v1",
        "rl.run.pause-requested.v1",
        "rl.run.resume-requested.v1",
        "rl.run.cancel-requested.v1",
        "rl.run.started.v1",
        "rl.run.metric-sampled.v1",
        "rl.run.episode-completed.v1",
        "rl.run.paused.v1",
        "rl.run.completed.v1",
        "rl.run.failed.v1",
        "rl.run.cancelled.v1",
    }
)


class MessageValidationError(ValueError):
    pass


@dataclass(frozen=True, slots=True)
class MessageEnvelope:
    event_id: str
    message_type: str
    schema_version: str
    occurred_at: str
    correlation_id: str
    causation_id: str
    run_id: str
    attempt: int
    payload: dict[str, Any]

    @classmethod
    def from_mapping(cls, value: dict[str, Any]) -> MessageEnvelope:
        required = {
            "event_id",
            "message_type",
            "schema_version",
            "occurred_at",
            "correlation_id",
            "causation_id",
            "run_id",
            "attempt",
            "payload",
        }
        if set(value) != required:
            raise MessageValidationError("message fields differ from envelope schema 1.0")
        if value["message_type"] not in ALLOWED_MESSAGE_TYPES:
            raise MessageValidationError("message type is not allowlisted")
        if value["schema_version"] != "1.0":
            raise MessageValidationError("only message schema 1.0 is supported")
        for key in ("event_id", "correlation_id", "causation_id", "run_id"):
            try:
                UUID(str(value[key]))
            except ValueError as exc:
                raise MessageValidationError(f"{key} is not a UUID") from exc
        try:
            occurred_at = datetime.fromisoformat(
                str(value["occurred_at"]).replace("Z", "+00:00")
            )
        except ValueError as exc:
            raise MessageValidationError("occurred_at is not ISO-8601") from exc
        if occurred_at.tzinfo is None:
            raise MessageValidationError("occurred_at must include a timezone")
        if not isinstance(value["attempt"], int) or not 1 <= value["attempt"] <= 5:
            raise MessageValidationError("attempt is outside the registered retry policy")
        if not isinstance(value["payload"], dict) or len(value["payload"]) > 20:
            raise MessageValidationError("payload is invalid or unbounded")
        return cls(**value)

    def as_dict(self) -> dict[str, Any]:
        return {
            "event_id": self.event_id,
            "message_type": self.message_type,
            "schema_version": self.schema_version,
            "occurred_at": self.occurred_at,
            "correlation_id": self.correlation_id,
            "causation_id": self.causation_id,
            "run_id": self.run_id,
            "attempt": self.attempt,
            "payload": self.payload,
        }


def new_event(
    source: MessageEnvelope,
    message_type: str,
    payload: dict[str, Any],
) -> MessageEnvelope:
    if message_type not in ALLOWED_MESSAGE_TYPES:
        raise MessageValidationError("message type is not allowlisted")
    return MessageEnvelope(
        event_id=str(uuid4()),
        message_type=message_type,
        schema_version="1.0",
        occurred_at=datetime.now(UTC).isoformat().replace("+00:00", "Z"),
        correlation_id=source.correlation_id,
        causation_id=source.event_id,
        run_id=source.run_id,
        attempt=source.attempt,
        payload=payload,
    )
