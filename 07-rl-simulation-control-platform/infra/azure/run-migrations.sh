#!/bin/sh
set -eu

: "${DATABASE_URL_DIRECT:?DATABASE_URL_DIRECT is required}"

psql "$DATABASE_URL_DIRECT" -v ON_ERROR_STOP=1 <<'SQL'
CREATE TABLE IF NOT EXISTS release_migration_checksums (
  version TEXT PRIMARY KEY,
  sha256 CHAR(64) NOT NULL,
  verified_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
SQL

for migration in /migrations/*.up.sql; do
  version="$(basename "$migration")"
  digest="$(sha256sum "$migration" | awk '{print $1}')"
  applied_digest="$(
    psql "$DATABASE_URL_DIRECT" -v ON_ERROR_STOP=1 -At \
      -v version="$version" <<'SQL'
SELECT sha256
FROM release_migration_checksums
WHERE version = :'version';
SQL
  )"

  if [ -n "$applied_digest" ]; then
    if [ "$applied_digest" != "$digest" ]; then
      echo "Migration checksum mismatch: $version" >&2
      exit 1
    fi
    echo "Already applied: $version"
    continue
  fi

  echo "Applying: $version"
  psql "$DATABASE_URL_DIRECT" -v ON_ERROR_STOP=1 -f "$migration"
  psql "$DATABASE_URL_DIRECT" -v ON_ERROR_STOP=1 \
    -v version="$version" -v digest="$digest" <<'SQL'
INSERT INTO release_migration_checksums (version, sha256)
VALUES (:'version', :'digest')
ON CONFLICT (version) DO UPDATE
SET sha256 = EXCLUDED.sha256, verified_at = NOW();
SQL
done

echo "OK - Neon schema migrations are current"
