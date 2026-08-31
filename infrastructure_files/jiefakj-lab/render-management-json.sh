#!/bin/sh
# Render management.json from management.json.example by substituting
# auth.example.com with AUTH_DOMAIN from the env file.
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname "$0")" && pwd)

if [ "${1:-}" != "" ]; then
	ENV_FILE=$1
fi
ENV_FILE="${ENV_FILE:-$SCRIPT_DIR/.env}"
OUT="${OUT:-$SCRIPT_DIR/management.json}"
EXAMPLE="${EXAMPLE:-$SCRIPT_DIR/management.json.example}"

if [ ! -f "$ENV_FILE" ]; then
	echo "env file not found: $ENV_FILE" >&2
	exit 1
fi
if [ ! -f "$EXAMPLE" ]; then
	echo "example not found: $EXAMPLE" >&2
	exit 1
fi

set -a
# shellcheck disable=SC1090
. "$ENV_FILE"
set +a

: "${AUTH_DOMAIN:?AUTH_DOMAIN must be set}"

sed "s|auth\\.example\\.com|${AUTH_DOMAIN}|g" "$EXAMPLE" >"$OUT"
