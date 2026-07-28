#!/usr/bin/env bash
set -euo pipefail

SERVICE=sub2api
USA0_ROOT=${USA0_ROOT:-/opt/usa0}
ASSET_DIR=${ASSET_DIR:-$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)}
ENV_FILE=${ENV_FILE:-$USA0_ROOT/config/$SERVICE.env}
export ENV_FILE
BACKUP_ROOT=${BACKUP_ROOT:-$USA0_ROOT/backups/$SERVICE}
DEPLOYMENT_ROOT=${DEPLOYMENT_ROOT:-$USA0_ROOT/deployments}

die() { echo "error: $*" >&2; exit 1; }
require() { command -v "$1" >/dev/null 2>&1 || die "$1 is required"; }
compose() { docker compose --env-file "$ENV_FILE" -f "$ASSET_DIR/compose.yaml" "${@:1}"; }

validate_env() {
  [[ -r $ENV_FILE ]] || die "missing $ENV_FILE"
  [[ $(stat -c %a "$ENV_FILE") == 600 ]] || die "$ENV_FILE must have mode 0600"
  if grep -Eq '=(replace-with|change_this|.*sha256:replace)' "$ENV_FILE"; then
    die "$ENV_FILE still contains placeholders"
  fi
  for key in APP_IMAGE POSTGRES_IMAGE REDIS_IMAGE POSTGRES_USER POSTGRES_PASSWORD POSTGRES_DB REDIS_PASSWORD ADMIN_EMAIL JWT_SECRET TOTP_ENCRYPTION_KEY DISTRIBUTION_TRACKING_HASH_SECRETS; do
    grep -Eq "^${key}=.+" "$ENV_FILE" || die "missing required value: $key"
  done
}

set_env_value() {
  local key=$1 value=$2 temp
  temp=$(mktemp "${ENV_FILE}.XXXXXX")
  awk -v key="$key" -v value="$value" 'BEGIN{found=0} index($0,key "=")==1{print key "=" value; found=1; next} {print} END{if(!found) print key "=" value}' "$ENV_FILE" >"$temp"
  chmod 600 "$temp"
  mv "$temp" "$ENV_FILE"
}

current_image() { sed -n 's/^APP_IMAGE=//p' "$ENV_FILE" | tail -n1; }
