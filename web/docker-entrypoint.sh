#!/bin/sh
# Writes build/env.js from the container environment before serving.
#
# Create React App inlines REACT_APP_* at build time, so a published image is
# otherwise frozen with whatever the CI runner had. Emitting a small script the
# page loads before the bundle gives us per-deployment config from one image.
set -eu

BUILD_DIR="${BUILD_DIR:-/app/build}"
ENV_FILE="${BUILD_DIR}/env.js"

# Keys the frontend understands. Unset keys are omitted entirely so that
# src/config.js falls through to its own defaults.
KEYS="API_URL MAX_PODS REFRESH_INTERVAL_MS HEAT_THRESHOLD SHOW_LABELS SHOW_POD_COUNT SHOW_TOOLTIPS SHOW_FIRE"

escape() {
  printf '%s' "$1" | sed -e 's/\\/\\\\/g' -e 's/"/\\"/g'
}

{
  printf 'window.__KUBE_CATS_ENV__ = {\n'
  for key in $KEYS; do
    value="$(printenv "$key" 2>/dev/null || true)"
    [ -n "$value" ] || continue
    printf '  "%s": "%s",\n' "$key" "$(escape "$value")"
  done
  printf '};\n'
} > "$ENV_FILE"

echo "kube-cats: wrote runtime config to $ENV_FILE"
cat "$ENV_FILE"

exec "$@"
