#!/usr/bin/env bash
set -euo pipefail
source "$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)/lib.sh"
version=${1:-}
target_image=${2:-}
[[ $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]] || die "invalid v-prefixed semantic version"
[[ $target_image =~ ^ghcr\.io/0xhyperia/sub2api@sha256:[0-9a-f]{64}$ ]] || die "target image must be the approved Sub2API digest"
require docker
validate_env
old_image=$(current_image)
APP_IMAGE=$target_image compose config --quiet
docker manifest inspect "$target_image" >/dev/null 2>&1 || die "image manifest unavailable: $target_image"
if [[ ${DRY_RUN:-false} == true ]]; then
  echo "dry-run passed: $SERVICE $version $old_image -> $target_image"
  exit 0
fi
initialize=${INITIALIZE:-false}
if compose ps --status running postgres | grep -q postgres; then
  backup=$("$ASSET_DIR/backup.sh")
elif [[ $initialize == true ]]; then
  set_env_value APP_IMAGE "$target_image"
  compose up -d postgres redis
  backup=none-initial-deployment
else
  die "postgres is not running; initialize infrastructure in a maintenance window"
fi
mkdir -p "$DEPLOYMENT_ROOT"; chmod 700 "$DEPLOYMENT_ROOT"
timestamp=$(date -u +%Y%m%dT%H%M%SZ)
record="$DEPLOYMENT_ROOT/${SERVICE}_${timestamp}.json"
set_env_value APP_IMAGE "$target_image"
status=failed
if docker pull "$target_image" && compose up -d --no-deps app; then
  for _ in $(seq 1 60); do "$ASSET_DIR/smoke.sh" internal >/dev/null 2>&1 && status=deployed && break; sleep 2; done
  [[ $status == deployed ]] && "$ASSET_DIR/smoke.sh" public || status=failed
fi
if [[ $status != deployed && $initialize != true ]]; then
  set_env_value APP_IMAGE "$old_image"
  compose up -d --no-deps app || true
fi
printf '{"service":"%s","version":"%s","source_commit":"%s","deploy_commit":"%s","previous_image":"%s","image":"%s","backup":"%s","status":"%s"}\n' \
  "$SERVICE" "$version" "${SOURCE_COMMIT:-unknown}" "${DEPLOY_COMMIT:-unknown}" "$old_image" "$target_image" "$backup" "$status" >"$record"
chmod 600 "$record"
if [[ $status != deployed ]]; then
  [[ $initialize == true ]] && die "initial deployment failed; data containers were left for diagnosis"
  die "deployment failed; application image restored; database backup: $backup"
fi
echo "deployment passed: record=$record backup=$backup"
