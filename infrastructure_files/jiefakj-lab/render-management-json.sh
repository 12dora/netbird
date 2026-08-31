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

if [ -z "${AUTH_DOMAIN:-}" ]; then
	AUTH_DOMAIN=$(sed -n 's/^AUTH_DOMAIN=//p' "$ENV_FILE" | tail -n 1)
fi

case "$AUTH_DOMAIN" in
	'"'*'"')
		AUTH_DOMAIN=${AUTH_DOMAIN#\"}
		AUTH_DOMAIN=${AUTH_DOMAIN%\"}
		;;
	"'"*"'")
		AUTH_DOMAIN=${AUTH_DOMAIN#\'}
		AUTH_DOMAIN=${AUTH_DOMAIN%\'}
		;;
esac

case "$AUTH_DOMAIN" in
	'' | *[!A-Za-z0-9.-]*)
		echo "invalid AUTH_DOMAIN: ${AUTH_DOMAIN}" >&2
		exit 1
		;;
esac

if ! sed "s|auth\\.example\\.com|${AUTH_DOMAIN}|g" "$EXAMPLE" >"$OUT.tmp"; then
	rm -f "$OUT.tmp"
	exit 1
fi
mv "$OUT.tmp" "$OUT"
