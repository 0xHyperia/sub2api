#!/usr/bin/env bash
set -euo pipefail
source "$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)/lib.sh"
require docker
validate_env
mkdir -p "$BACKUP_ROOT"
chmod 700 "$BACKUP_ROOT"
timestamp=$(date -u +%Y%m%dT%H%M%SZ)
dump="$BACKUP_ROOT/${SERVICE}_${timestamp}.dump"
partial="$dump.partial"
trap 'rm -f -- "$partial"' EXIT
compose exec -T postgres sh -c 'exec pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc' >"$partial"
test -s "$partial" || die "database backup is empty"
mv "$partial" "$dump"
chmod 600 "$dump"
echo "$dump"
