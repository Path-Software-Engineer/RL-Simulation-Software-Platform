from __future__ import annotations

import json
import logging
import os
import time
from pathlib import Path
from typing import Any

import redis
from redis.exceptions import ResponseError

from .dqn import DQNConfig, train_dqn
from .messages import MessageEnvelope, MessageValidationError, new_event
from .policies import PolicyRegistry, PolicyValidationError
from .runner import RunRequest, run_episode

LOGGER = logging.getLogger("rl_runner")


def configure_logging() -> None:
    logging.basicConfig(level=os.getenv("LOG_LEVEL", "INFO"), format="%(message)s")


class RunnerConsumer:
    def __init__(self, client: redis.Redis, artifact_root: Path) -> None:
        self.client = client
        self.registry = PolicyRegistry(artifact_root)
        self.command_stream = os.getenv("COMMAND_STREAM", "rl.commands.v1")
        self.event_stream = os.getenv("EVENT_STREAM", "rl.events.v1")
        self.dead_letter_stream = os.getenv("DEAD_LETTER_STREAM", "rl.dead-letter.v1")
        self.group = os.getenv("RUNNER_CONSUMER_GROUP", "rl-runner-v1")
        self.consumer = os.getenv("HOSTNAME", "runner-local")[:64]
        self.step_delay_seconds = min(
            max(int(os.getenv("RUNNER_STEP_DELAY_MS", "120")), 0), 1000
        ) / 1000
        self.dqn_step_delay_seconds = min(
            max(int(os.getenv("DQN_STEP_DELAY_MS", "10")), 0), 250
        ) / 1000

    def ensure_group(self) -> None:
        try:
            self.client.xgroup_create(self.command_stream, self.group, id="0", mkstream=True)
        except ResponseError as exc:
            if "BUSYGROUP" not in str(exc):
                raise

    def run_forever(self) -> None:
        self.ensure_group()
        LOGGER.info(json.dumps({"event": "runner.ready", "consumer": self.consumer}))
        while True:
            batches = self.client.xreadgroup(
                self.group,
                self.consumer,
                {self.command_stream: ">"},
                count=1,
                block=5000,
            )
            for _, messages in batches:
                for message_id, fields in messages:
                    self._process(message_id, fields)

    def _process(self, message_id: str, fields: dict[str, str]) -> None:
        raw = fields.get("data")
        if raw is None:
            self._dead_letter(message_id, "missing data field", fields)
            return
        try:
            envelope = MessageEnvelope.from_mapping(json.loads(raw))
            if self.client.sismember(f"{self.group}:processed", envelope.event_id):
                self.client.xack(self.command_stream, self.group, message_id)
                return
            if envelope.message_type == "rl.run.requested.v1":
                self._execute(envelope)
            elif envelope.message_type in {
                "rl.run.pause-requested.v1",
                "rl.run.resume-requested.v1",
                "rl.run.cancel-requested.v1",
            }:
                self._apply_control(envelope)
            else:
                raise MessageValidationError("message type is not valid on the command stream")
            self.client.sadd(f"{self.group}:processed", envelope.event_id)
            self.client.expire(f"{self.group}:processed", 604800)
            self.client.xack(self.command_stream, self.group, message_id)
        except (
            json.JSONDecodeError,
            KeyError,
            MessageValidationError,
            OverflowError,
            PolicyValidationError,
            TypeError,
            ValueError,
        ) as exc:
            self._dead_letter(message_id, str(exc), {"data": raw})

    def _apply_control(self, envelope: MessageEnvelope) -> None:
        state_by_message = {
            "rl.run.pause-requested.v1": "pause",
            "rl.run.resume-requested.v1": "run",
            "rl.run.cancel-requested.v1": "cancel",
        }
        self.client.set(
            f"rl.control:{envelope.run_id}",
            state_by_message[envelope.message_type],
            ex=900,
        )

    def _execute(self, envelope: MessageEnvelope) -> None:
        payload = envelope.payload
        policy = self.registry.load(str(payload["policy_id"]), str(payload["policy_sha256"]))
        request = RunRequest(
            run_id=envelope.run_id,
            environment_id=str(payload["environment_id"]),
            environment_version=str(payload["environment_version"]),
            policy_id=str(payload["policy_id"]),
            policy_sha256=str(payload["policy_sha256"]),
            seed=int(payload["seed"]),
            max_steps=int(payload["max_steps"]),
        )
        request.validate(policy)
        self._publish(new_event(envelope, "rl.run.started.v1", {"started_by": self.consumer}))
        if policy.algorithm == "dqn":
            self._execute_dqn(envelope, request, policy.training_config)
            return
        result = run_episode(request, policy, control=self._cooperative_control(envelope))
        self._publish(new_event(envelope, "rl.run.episode-completed.v1", result.as_payload()))
        metrics = (
            ("episode_reward", result.total_reward, "reward"),
            ("episode_steps", result.step_count, "steps"),
            ("collisions", result.collisions, "count"),
        )
        for metric, value, unit in metrics:
            self._publish(
                new_event(
                    envelope,
                    "rl.run.metric-sampled.v1",
                    {"metric": metric, "value": value, "unit": unit, "step": 1},
                )
            )
        terminal_type = (
            "rl.run.cancelled.v1" if result.status == "cancelled" else "rl.run.completed.v1"
        )
        self._publish(
            new_event(
                envelope,
                terminal_type,
                {"status": result.status, "terminal_reason": result.terminal_reason},
            )
        )

    def _execute_dqn(
        self,
        envelope: MessageEnvelope,
        request: RunRequest,
        raw_config: dict[str, Any] | None,
    ) -> None:
        if raw_config is None:
            raise ValueError("DQN training configuration is missing")

        def publish_episode(episode, metrics) -> None:
            self._publish(
                new_event(envelope, "rl.run.episode-completed.v1", episode.as_payload())
            )
            for metric in metrics:
                self._publish(
                    new_event(
                        envelope,
                        "rl.run.metric-sampled.v1",
                        {
                            "metric": metric.metric,
                            "value": metric.value,
                            "unit": metric.unit,
                            "step": metric.step,
                        },
                    )
                )

        result = train_dqn(
            request,
            DQNConfig.from_mapping(raw_config),
            control=self._cooperative_control(
                envelope, delay_seconds=self.dqn_step_delay_seconds
            ),
            on_episode=publish_episode,
        )
        terminal_type = (
            "rl.run.cancelled.v1"
            if result.status == "cancelled"
            else "rl.run.completed.v1"
        )
        self._publish(
            new_event(
                envelope,
                terminal_type,
                {"status": result.status, "terminal_reason": result.terminal_reason},
            )
        )

    def _cooperative_control(
        self, envelope: MessageEnvelope, *, delay_seconds: float | None = None
    ):
        pause_announced = False
        effective_delay = self.step_delay_seconds if delay_seconds is None else delay_seconds

        def check() -> str:
            nonlocal pause_announced
            if effective_delay:
                time.sleep(effective_delay)
            state = self.client.get(f"rl.control:{envelope.run_id}") or "run"
            if state == "pause" and not pause_announced:
                self._publish(
                    new_event(envelope, "rl.run.paused.v1", {"cooperative": True})
                )
                pause_announced = True
            while state == "pause":
                time.sleep(0.1)
                state = self.client.get(f"rl.control:{envelope.run_id}") or "run"
            if pause_announced and state == "run":
                self._publish(
                    new_event(envelope, "rl.run.started.v1", {"resumed": True})
                )
                pause_announced = False
            return "cancel" if state == "cancel" else "run"

        return check

    def _publish(self, envelope: MessageEnvelope) -> None:
        self.client.xadd(
            self.event_stream,
            {"data": json.dumps(envelope.as_dict(), separators=(",", ":"))},
            maxlen=10000,
            approximate=True,
        )

    def _dead_letter(self, message_id: str, error: str, fields: dict[str, Any]) -> None:
        self.client.xadd(
            self.dead_letter_stream,
            {"source_id": message_id, "error": error[:500], "payload": json.dumps(fields)[:4000]},
            maxlen=1000,
            approximate=True,
        )
        self.client.xack(self.command_stream, self.group, message_id)
        LOGGER.error(json.dumps({"event": "runner.message_rejected", "error": error}))


def main() -> None:
    configure_logging()
    client = redis.Redis.from_url(
        os.getenv("REDIS_URL", "redis://127.0.0.1:6379/0"), decode_responses=True
    )
    artifact_root = Path(
        os.getenv("POLICY_ARTIFACT_ROOT", "/workspace/artifacts/policies")
    )
    RunnerConsumer(client, artifact_root).run_forever()


if __name__ == "__main__":
    main()
