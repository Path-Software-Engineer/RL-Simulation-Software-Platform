from __future__ import annotations

import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def load(relative: str) -> dict:
    return json.loads((ROOT / relative).read_text(encoding="utf-8"))


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def resolve_local_pointer(document: dict, reference: str) -> object:
    require(reference.startswith("#/"), f"unsupported external reference: {reference}")
    value: object = document
    for segment in reference[2:].split("/"):
        segment = segment.replace("~1", "/").replace("~0", "~")
        require(
            isinstance(value, dict) and segment in value,
            f"unresolved reference: {reference}",
        )
        value = value[segment]
    return value


def validate_local_references(document: dict) -> None:
    pending: list[object] = [document]
    while pending:
        current = pending.pop()
        if isinstance(current, dict):
            reference = current.get("$ref")
            if isinstance(reference, str) and reference.startswith("#/"):
                resolve_local_pointer(document, reference)
            pending.extend(current.values())
        elif isinstance(current, list):
            pending.extend(current)


def validate_create_run(instance: dict, schema: dict) -> bool:
    if set(instance) - set(schema["properties"]):
        return False
    if any(field not in instance for field in schema["required"]):
        return False
    return (
        isinstance(instance["environmentId"], str)
        and isinstance(instance["policyId"], str)
        and isinstance(instance["seed"], int)
        and 0 <= instance["seed"] <= 2_147_483_647
        and isinstance(instance["maxSteps"], int)
        and 1 <= instance["maxSteps"] <= 200
    )


def main() -> None:
    openapi = load("contracts/http/openapi.json")
    asyncapi = load("contracts/events/asyncapi.json")
    envelope = load("contracts/events/schemas/message-envelope-v1.schema.json")
    request = load("contracts/events/schemas/run-requested-v1.schema.json")
    run_message = load("contracts/events/schemas/run-requested-message-v1.schema.json")
    episode_message = load("contracts/events/schemas/episode-completed-message-v1.schema.json")
    valid = load("contracts/http/examples/create-run.valid.json")
    invalid = load("contracts/http/examples/create-run.invalid.json")
    message = load("contracts/events/examples/run-requested.valid.json")

    require(openapi.get("openapi") == "3.1.0", "OpenAPI must remain at 3.1.0")
    require(asyncapi.get("asyncapi") == "3.0.0", "AsyncAPI must remain at 3.0.0")
    require(openapi["info"]["version"] == "0.3.0", "OpenAPI Sprint 3 version is missing")
    require(asyncapi["info"]["version"] == "0.3.0", "AsyncAPI Sprint 3 version is missing")
    validate_local_references(openapi)
    validate_local_references(asyncapi)
    require(
        openapi.get("security") == [{"operatorBearer": []}],
        "REST operator auth is not registered",
    )
    require(
        "operatorBearer" in openapi["components"]["securitySchemes"],
        "bearer scheme is missing",
    )
    websocket = openapi["paths"]["/ws/v1/runs/{id}"]["get"]
    require(
        websocket.get("x-websocket-auth", {}).get("subprotocols")
        == ["rl-run-v1", "operator-token"],
        "WebSocket auth is not registered",
    )
    required_paths = {
        "/api/v1/environments/{id}",
        "/api/v1/policies/{id}",
        "/api/v1/training-runs",
        "/api/v1/training-runs/{id}/episodes",
        "/api/v1/episodes/{id}/transitions",
        "/ws/v1/runs/{id}",
    }
    require(required_paths <= set(openapi["paths"]), "OpenAPI is missing a Sprint 1 route")
    policy_algorithms = set(
        openapi["components"]["schemas"]["Policy"]["properties"]["algorithm"]["enum"]
    )
    require("world-model" in policy_algorithms, "world-model policy is not registered")
    transition_properties = set(
        openapi["components"]["schemas"]["Transition"]["properties"]
    )
    require(
        {
            "predictedState",
            "predictedNextState",
            "stepError",
            "accumulatedError",
            "modelVersion",
        }
        <= transition_properties,
        "Sprint 3 transition evidence is missing",
    )
    create_schema = openapi["components"]["schemas"]["CreateRunRequest"]
    require(validate_create_run(valid, create_schema), "valid HTTP example was rejected")
    require(
        not validate_create_run(invalid, create_schema),
        "invalid HTTP example was accepted",
    )
    require(
        envelope["$id"].endswith("message-envelope-v1.schema.json"),
        "envelope id is not versioned",
    )
    require(
        request["$id"].endswith("run-requested-v1.schema.json"),
        "request id is not versioned",
    )
    require(
        len(run_message["allOf"]) == 2,
        "run message does not compose envelope and payload",
    )
    require(
        len(episode_message["allOf"]) == 2,
        "episode message does not compose envelope and payload",
    )
    message_names = {item["name"] for item in asyncapi["components"]["messages"].values()}
    require(
        "rl.run.control-requested.v1" not in message_names,
        "AsyncAPI contains a non-runtime message type",
    )
    require(set(message) == set(envelope["required"]), "message example differs from envelope")
    require(message["message_type"] == "rl.run.requested.v1", "message type is not allowlisted")
    print("OK - OpenAPI, AsyncAPI, schemas and examples are structurally valid")


if __name__ == "__main__":
    main()
