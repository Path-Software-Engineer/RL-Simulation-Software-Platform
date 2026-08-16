from __future__ import annotations

import json
from copy import deepcopy
from pathlib import Path

from jsonschema import FormatChecker, ValidationError
from jsonschema.validators import validator_for
from openapi_spec_validator import validate
from referencing import Registry, Resource

ROOT = Path(__file__).resolve().parents[1]
SCHEMA_ROOT = ROOT / "contracts" / "events" / "schemas"


def load(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def main() -> None:
    openapi = load(ROOT / "contracts" / "http" / "openapi.json")
    validate(openapi)

    create_run_schema = openapi["components"]["schemas"]["CreateRunRequest"]
    create_run_validator = validator_for(create_run_schema)(create_run_schema)
    valid_http_example = load(
        ROOT / "contracts" / "http" / "examples" / "create-run.valid.json"
    )
    invalid_http_example = load(
        ROOT / "contracts" / "http" / "examples" / "create-run.invalid.json"
    )
    create_run_validator.validate(valid_http_example)
    try:
        create_run_validator.validate(invalid_http_example)
    except ValidationError:
        pass
    else:
        raise SystemExit("invalid HTTP example was accepted by the OpenAPI schema")

    schemas = [load(path) for path in sorted(SCHEMA_ROOT.glob("*.schema.json"))]
    registry = Registry().with_resources(
        (schema["$id"], Resource.from_contents(schema)) for schema in schemas
    )
    for schema in schemas:
        validator_for(schema).check_schema(schema)

    run_schema = next(
        schema
        for schema in schemas
        if schema["$id"].endswith("run-requested-message-v1.schema.json")
    )
    run_example = load(
        ROOT / "contracts" / "events" / "examples" / "run-requested.valid.json"
    )
    run_validator = validator_for(run_schema)(
        run_schema,
        registry=registry,
        format_checker=FormatChecker(),
    )
    run_validator.validate(run_example)

    invalid = deepcopy(run_example)
    invalid["payload"]["max_steps"] = 201
    try:
        run_validator.validate(invalid)
    except ValidationError:
        pass
    else:
        raise SystemExit("invalid event example was accepted by JSON Schema")

    print("OK - formal OpenAPI 3.1 and JSON Schema 2020-12 validation passed")


if __name__ == "__main__":
    main()
