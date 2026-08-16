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
        if set(payload["actionByState"].values()) - {"up", "right", "down", "left", "blocked"}:
            raise SystemExit(f"policy contains an arbitrary action: {path.name}")
    environment = ROOT / "artifacts/manifests/gridworld-environment-v1.json"
    if digest(environment) != "92da1f86009908279cd2cd9d4323056f0a83fb6bdd984dd67fcb8dd7ea28fb49":
        raise SystemExit("environment manifest hash mismatch")
    print(f"OK - {len(manifest['policies'])} policy artifacts and environment manifest verified")


if __name__ == "__main__":
    main()
