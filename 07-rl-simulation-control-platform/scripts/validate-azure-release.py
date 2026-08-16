from __future__ import annotations

import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def read(relative: str) -> str:
    path = ROOT / relative
    if not path.is_file():
        raise SystemExit(f"missing Azure release file: {relative}")
    return path.read_text(encoding="utf-8")


def require(text: str, tokens: tuple[str, ...], label: str) -> None:
    missing = [token for token in tokens if token not in text]
    if missing:
        raise SystemExit(f"{label} is missing required declarations: {', '.join(missing)}")


def main() -> None:
    foundation = read("infra/azure/foundation.bicep")
    workloads = read("infra/azure/workloads.bicep")
    workload_params = read("infra/azure/workloads.bicepparam")
    migration = read("infra/azure/run-migrations.sh")
    migration_image = read("infra/docker/migrate.Dockerfile")
    gateway_image = read("infra/docker/gateway.Dockerfile")
    gateway_config = read("infra/azure/Caddyfile")
    deploy = read("scripts/deploy-azure.ps1")
    cli_installer = read("scripts/install-azure-cli-current-user.ps1")
    smoke = read("scripts/smoke-test-release.ps1")
    compose = read("docker-compose.yml")

    require(
        foundation,
        (
            "Microsoft.App/managedEnvironments@",
            "Microsoft.ContainerRegistry/registries@",
            "Microsoft.ManagedIdentity/userAssignedIdentities@",
            "Microsoft.Authorization/roleAssignments@",
            "destination: 'none'",
            "name: 'Standard'",
            "adminUserEnabled: false",
            "costProfile: 'free-grant-scale-to-zero'",
        ),
        "Azure foundation",
    )
    forbidden_foundation = (
        "Microsoft.Cache/",
        "Microsoft.OperationalInsights/",
        "Balanced_B0",
        "name: 'Basic'",
        "log-analytics",
    )
    if any(token in foundation for token in forbidden_foundation):
        raise SystemExit("Azure foundation contains a billable or non-free-grant service")
    require(
        workloads,
        (
            "Microsoft.App/jobs@",
            "Microsoft.App/containerApps@",
            "param databaseUrl string",
            "param databaseUrlDirect string",
            "param operatorToken string",
            "secretRef: 'database-url'",
            "secretRef: 'database-url-direct'",
            "secretRef: 'operator-token'",
            "image: 'redis:8.2.1-alpine'",
            "value: 'redis://127.0.0.1:6379/0'",
            "value: 'wss://${appFqdn}'",
            "minReplicas: 0",
            "maxReplicas: 1",
            "targetPort: 8088",
        ),
        "Azure workloads",
    )
    secure_params = re.findall(r"@secure\(\)\s+(?:@[^\n]+\s+)*param\s+(\w+)\s+string", workloads)
    if set(secure_params) != {"databaseUrl", "databaseUrlDirect", "operatorToken"}:
        raise SystemExit(
            "all three workload credentials must be declared as secure Bicep parameters"
        )
    secure_output = r"output\s+\w+\s+string\s*=\s*(databaseUrl|databaseUrlDirect|operatorToken)"
    if re.search(secure_output, workloads):
        raise SystemExit("a secure workload parameter must never be emitted as a Bicep output")
    if "minReplicas: 1" in workloads or "Microsoft.Cache/" in workloads:
        raise SystemExit("Azure workloads must scale to zero and must not provision managed Redis")
    require(
        workload_params,
        (
            "using './workloads.bicep'",
            "readEnvironmentVariable('NEON_DATABASE_URL')",
            "readEnvironmentVariable('NEON_DATABASE_URL_DIRECT')",
            "readEnvironmentVariable('OPERATOR_TOKEN')",
        ),
        "secret-safe workload parameters",
    )

    require(
        migration,
        (
            'set -eu',
            'DATABASE_URL_DIRECT',
            'release_migration_checksums',
            'sha256sum',
            'ON_ERROR_STOP=1',
            'Already applied:',
            'Migration checksum mismatch:',
            "WHERE version = :'version';",
            "VALUES (:'version', :'digest')",
            "<<'SQL'",
        ),
        "Neon migration runner",
    )
    if re.search(r'-c\s+"[^"]*:\'(version|digest)\'', migration):
        raise SystemExit("psql variables must be expanded from an input script, not a -c argument")
    migrations = sorted((ROOT / "database/migrations").glob("*.up.sql"))
    if [path.name for path in migrations] != [
        "0001_sprint_01_gridworld.up.sql",
        "0002_sprint_02_dqn_dashboard.up.sql",
        "0003_sprint_03_world_model_rollout_viewer.up.sql",
    ]:
        raise SystemExit(
            "the release migration set differs from the three versioned sprint migrations"
        )
    for path in migrations:
        sql = path.read_text(encoding="utf-8")
        if len(re.findall(r"(?im)^\s*BEGIN;\s*$", sql)) != 1:
            raise SystemExit(f"{path.name} must contain exactly one BEGIN statement")
        if len(re.findall(r"(?im)^\s*COMMIT;\s*$", sql)) != 1:
            raise SystemExit(f"{path.name} must contain exactly one COMMIT statement")

    require(
        migration_image,
        (
            "FROM timescale/timescaledb:2.29.0-pg17",
            "COPY database/migrations/ /migrations/",
            "USER postgres",
            "ENTRYPOINT [\"/usr/local/bin/run-migrations\"]",
        ),
        "migration image",
    )
    require(
        gateway_image,
        (
            "FROM caddy:2.10.2-alpine",
            "COPY infra/azure/Caddyfile /etc/caddy/Caddyfile",
        ),
        "same-origin gateway image",
    )
    require(
        gateway_config,
        (
            "auto_https off",
            "reverse_proxy 127.0.0.1:8080",
            "reverse_proxy 127.0.0.1:3000",
            "/ws/*",
        ),
        "same-origin gateway configuration",
    )
    require(
        deploy,
        (
            'Read-RequiredSecret "NEON_DATABASE_URL"',
            'Read-RequiredSecret "NEON_DATABASE_URL_DIRECT"',
            'Read-RequiredSecret "OPERATOR_TOKEN"',
            '"containerapp", "job", "start"',
            'Dockerfile = "infra/docker/gateway.Dockerfile"',
            '"--parameters", "infra/azure/workloads.bicepparam"',
            'smoke-test-release.ps1',
            'git status --porcelain',
        ),
        "Azure deployment script",
    )
    forbidden_secret_arguments = (
        '"databaseUrl=$DatabaseUrl"',
        '"databaseUrlDirect=$DatabaseUrlDirect"',
        '"operatorToken=$OperatorToken"',
    )
    if any(token in deploy for token in forbidden_secret_arguments):
        raise SystemExit("deployment secrets must not be passed through Azure CLI arguments")
    require(
        cli_installer,
        (
            'Join-Path $env:LOCALAPPDATA "Programs\\AzureCLI\\$Version"',
            "https://azcliprod.blob.core.windows.net/zip/azure-cli-$Version-x64.zip",
            "Invoke-WebRequest",
            "Expand-Archive",
            '[Environment]::SetEnvironmentVariable(',
            '"Path",',
            '"User"',
            "az.cmd",
        ),
        "current-user Azure CLI installer",
    )
    require(
        smoke,
        (
            "^https://",
            "/health/ready",
            "/openapi.json",
            "smoke-test.ps1",
            "smoke-test-sprint-02.ps1",
            "smoke-test-sprint-03.ps1",
        ),
        "release smoke",
    )
    require(
        compose,
        (
            "release-migrate:",
            'profiles: ["release-tools"]',
            "dockerfile: infra/docker/migrate.Dockerfile",
            "DATABASE_URL_DIRECT:",
        ),
        "local release migration parity service",
    )

    release_sources = (
        foundation,
        workloads,
        workload_params,
        migration,
        migration_image,
        gateway_image,
        gateway_config,
        deploy,
        cli_installer,
        smoke,
    )
    if any(":latest" in source for source in release_sources):
        raise SystemExit("Azure release sources must not reference mutable latest image tags")
    if re.search(r"(?i)write-(host|output).*?(databaseurl|rediskey|operatorToken)", deploy):
        raise SystemExit("the deployment script must not print release credentials")

    print(
        "OK - zero-cost scale-to-zero Azure Container Apps and Neon assets are "
        "structurally complete"
    )


if __name__ == "__main__":
    main()
