from pathlib import Path

from rl_runner.consumer import RunnerConsumer
from rl_runner.messages import MessageEnvelope


class StubRedis:
    def __init__(self) -> None:
        self.states = ["pause", "run"]

    def get(self, _key: str) -> str:
        return self.states.pop(0)


def test_pause_and_resume_are_confirmed_by_runner() -> None:
    consumer = RunnerConsumer(StubRedis(), Path("artifacts"))
    consumer.step_delay_seconds = 0
    published: list[str] = []
    consumer._publish = (  # type: ignore[method-assign]
        lambda event: published.append(event.message_type)
    )
    source = MessageEnvelope(
        event_id="11111111-1111-4111-8111-111111111111",
        message_type="rl.run.requested.v1",
        schema_version="1.0",
        occurred_at="2026-08-14T00:00:00Z",
        correlation_id="22222222-2222-4222-8222-222222222222",
        causation_id="11111111-1111-4111-8111-111111111111",
        run_id="33333333-3333-4333-8333-333333333333",
        attempt=1,
        payload={},
    )

    assert consumer._cooperative_control(source)() == "run"
    assert published == ["rl.run.paused.v1", "rl.run.started.v1"]
