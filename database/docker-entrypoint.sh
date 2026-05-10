#!/usr/bin/env bash
set -euo pipefail

# Re-apply idempotent schema after Postgres accepts connections (every container start).
# Official initdb.d hooks only run on first empty PGDATA; this wrapper covers restarts.

if [ "${1:-}" = 'postgres' ]; then
	docker-entrypoint.sh "$@" &
	pid=$!
	until pg_isready -q -h localhost -p "${PGPORT:-5432}" -U "${POSTGRES_USER:-postgres}"; do
		sleep 0.2
	done
	if ! psql -v ON_ERROR_STOP=1 -h localhost -p "${PGPORT:-5432}" -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-postgres}" -f /schema/db.sql; then
		kill -TERM "$pid" 2>/dev/null || true
		wait "$pid" 2>/dev/null || true
		exit 1
	fi
	wait "$pid"
else
	exec docker-entrypoint.sh "$@"
fi
