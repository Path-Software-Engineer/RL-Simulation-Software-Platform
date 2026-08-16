from __future__ import annotations

import hashlib
import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Final

from .environment import ACTIONS, Action, Coordinate

Q_LEARNING_ID: Final[str] = "22222222-2222-4222-8222-222222222201"
SARSA_ID: Final[str] = "22222222-2222-4222-8222-222222222202"
DQN_ID: Final[str] = "22222222-2222-4222-8222-222222222203"

EXPECTED_HASHES: Final[dict[str, str]] = {
    Q_LEARNING_ID: "22d5faf9a94fcd05fdf31d2a1429a1b8f02d4f394f6cb61b95cc26351060927f",
    SARSA_ID: "b7043771657ec37f235103485a7375a395431db0b2dc0e39d2070ca4e87aadcb",
    DQN_ID: "99902302d7119e631c23d982eba7a06b0135718a81780caa55e68d0732dccac7",
}

ARTIFACT_FILES: Final[dict[str, str]] = {
    Q_LEARNING_ID: "q-learning-gridworld-v1.json",
    SARSA_ID: "sarsa-gridworld-v1.json",
    DQN_ID: "dqn-gridworld-training-v1.json",
}


class PolicyValidationError(ValueError):
    pass


@dataclass(frozen=True, slots=True)
class PolicyArtifact:
    policy_id: str
    algorithm: str
    version: str
    environment_id: str
    sha256: str
    action_by_state: dict[str, str]
    training_config: dict[str, Any] | None

    def action_for(self, state: Coordinate) -> Action:
        value = self.action_by_state.get(state.key)
        if value not in ACTIONS:
            raise PolicyValidationError(f"no executable action for state {state.key}")
        return value


class PolicyRegistry:
    def __init__(self, artifact_root: Path) -> None:
        self._artifact_root = artifact_root.resolve()

    def load(self, policy_id: str, claimed_sha256: str) -> PolicyArtifact:
        file_name = ARTIFACT_FILES.get(policy_id)
        expected_hash = EXPECTED_HASHES.get(policy_id)
        if file_name is None or expected_hash is None:
            raise PolicyValidationError("policy is not allowlisted")
        if claimed_sha256 != expected_hash:
            raise PolicyValidationError("claimed policy hash differs from the registry")

        path = (self._artifact_root / file_name).resolve()
        if path.parent != self._artifact_root:
            raise PolicyValidationError("policy path escaped the artifact registry")
        content = path.read_bytes()
        actual_hash = hashlib.sha256(content).hexdigest()
        if actual_hash != expected_hash:
            raise PolicyValidationError("policy artifact integrity check failed")

        raw = json.loads(content)
        if raw.get("policyId") != policy_id or raw.get("version") != "1.0.0":
            raise PolicyValidationError("policy identity or version is incompatible")
        if raw.get("observationSpace") != "Discrete(36)":
            raise PolicyValidationError("policy observation space is incompatible")
        if tuple(raw.get("actionSpace", [])) != ACTIONS:
            raise PolicyValidationError("policy action space is incompatible")
        algorithm = str(raw.get("algorithm", ""))
        if algorithm not in {"q-learning", "sarsa", "dqn"}:
            raise PolicyValidationError("policy algorithm is not allowlisted")
        action_by_state = {
            str(key): str(value) for key, value in raw.get("actionByState", {}).items()
        }
        training_config = raw.get("training")
        if algorithm == "dqn" and not isinstance(training_config, dict):
            raise PolicyValidationError("DQN training configuration is missing")
        if algorithm != "dqn" and len(action_by_state) != 36:
            raise PolicyValidationError("tabular policy does not cover the registered state space")

        return PolicyArtifact(
            policy_id=policy_id,
            algorithm=algorithm,
            version=str(raw["version"]),
            environment_id=str(raw["environmentId"]),
            sha256=actual_hash,
            action_by_state=action_by_state,
            training_config=training_config if isinstance(training_config, dict) else None,
        )
