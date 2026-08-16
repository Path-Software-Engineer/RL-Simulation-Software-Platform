from __future__ import annotations

import hashlib
import json
from dataclasses import dataclass
from pathlib import Path
from typing import Final

from .environment import ACTIONS, Action, Coordinate

Q_LEARNING_ID: Final[str] = "22222222-2222-4222-8222-222222222201"
SARSA_ID: Final[str] = "22222222-2222-4222-8222-222222222202"

EXPECTED_HASHES: Final[dict[str, str]] = {
    Q_LEARNING_ID: "58d3eac12251df3a3e708eeb6544a056921db9d6f4b2f4930db4d85ae32f7a08",
    SARSA_ID: "88f759a4a95a4ad71fb68c162aeecc8f7b0c5a1443666d841b84cd6db1667fa3",
}

ARTIFACT_FILES: Final[dict[str, str]] = {
    Q_LEARNING_ID: "q-learning-gridworld-v1.json",
    SARSA_ID: "sarsa-gridworld-v1.json",
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

        return PolicyArtifact(
            policy_id=policy_id,
            algorithm=str(raw["algorithm"]),
            version=str(raw["version"]),
            environment_id=str(raw["environmentId"]),
            sha256=actual_hash,
            action_by_state={str(key): str(value) for key, value in raw["actionByState"].items()},
        )
