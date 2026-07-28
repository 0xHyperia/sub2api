#!/usr/bin/env bash
set -euo pipefail
source "$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)/lib.sh"
mode=${1:-internal}
validate_env
if [[ $mode == internal ]]; then
  compose exec -T app wget -q -T 5 -O /dev/null http://127.0.0.1:8080/health
  exit 0
fi
[[ $mode == public ]] || die "usage: smoke.sh internal|public"
require curl
set -a; source "$ENV_FILE"; set +a
origin=${PUBLIC_ORIGIN:-https://usa0.top}
curl --fail --silent --show-error "$origin/health" >/dev/null
if [[ -n ${SMOKE_ADMIN_EMAIL:-} && -n ${SMOKE_ADMIN_PASSWORD:-} ]]; then
  payload=$(printf '{"email":"%s","password":"%s"}' "$SMOKE_ADMIN_EMAIL" "$SMOKE_ADMIN_PASSWORD")
  curl --fail --silent --show-error -H 'content-type: application/json' --data "$payload" "$origin/api/v1/auth/login" >/dev/null
fi
if [[ -n ${SMOKE_API_KEY:-} && -n ${SMOKE_MODEL:-} ]]; then
  curl --fail --silent --show-error -H "authorization: Bearer $SMOKE_API_KEY" "$origin/v1/models" >/dev/null
  curl --fail --silent --show-error -H "authorization: Bearer $SMOKE_API_KEY" -H 'content-type: application/json' \
    --data "{\"model\":\"$SMOKE_MODEL\",\"messages\":[{\"role\":\"user\",\"content\":\"Reply OK\"}],\"stream\":true,\"max_tokens\":2}" \
    "$origin/v1/chat/completions" | grep -q '^data:'
fi
echo "sub2api $mode smoke passed"
