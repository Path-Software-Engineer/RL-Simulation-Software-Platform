from __future__ import annotations

import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main() -> None:
    manifest = json.loads(
        (ROOT / "artifacts/manifests/policy-artifacts-v1.json").read_text(encoding="utf-8")
    )
    for item in manifest["policies"]:
        path = ROOT / "artifacts" / item["uri"].removeprefix("artifact://")
        if not path.is_file():
            raise SystemExit(f"missing policy artifact: {path}")
        if digest(path) != item["sha256"]:
            raise SystemExit(f"policy hash mismatch: {path.name}")
        payload = json.loads(path.read_text(encoding="utf-8"))
        if payload["policyId"] != item["policyId"] or payload["version"] != item["version"]:
            raise SystemExit(f"policy identity mismatch: {path.name}")
        if set(payload.get("actionByState", {}).values()) - {
            "up",
            "right",
            "down",
            "left",
            "blocked",
        }:
            raise SystemExit(f"policy contains an arbitrary action: {path.name}")
        if payload["algorithm"] == "dqn":
            training = payload.get("training", {})
            required = {
                "episodes",
                "hiddenUnits",
                "gamma",
                "learningRate",
                "replayCapacity",
                "batchSize",
                "targetSyncSteps",
                "epsilonStart",
                "epsilonEnd",
                "epsilonDecayEpisodes",
                "movingAverageWindow",
            }
            if set(training) != required or not 1 <= training["episodes"] <= 100:
                raise SystemExit(f"DQN training profile is invalid: {path.name}")
    environment = ROOT / "artifacts/manifests/gridworld-environment-v1.json"
    if digest(environment) != "1b818490c0e1ab1b17753f0d089d547686035efb4c52f4e6bbbf0dfac23cbc43":
        raise SystemExit("environment manifest hash mismatch")
    print(f"OK - {len(manifest['policies'])} policy artifacts and environment manifest verified")


if __name__ == "__main__":
    main()
