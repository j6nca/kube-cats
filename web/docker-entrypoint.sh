#!/bin/sh
# Writes the runtime config script from the container environment before serving.
#
# Create React App inlines REACT_APP_* at build time, so a published image is
# otherwise frozen with whatever the CI runner had. Emitting a small script the
# page loads before the bundle gives us per-deployment config from one image.
#
# This writes into build/config/ rather than build/ itself so the target can be
# a mounted volume: the image ships read-only and root-owned, and deployments
# commonly run as a non-root uid with readOnlyRootFilesystem. Mount a writable
# volume (an emptyDir is enough) at /app/build/config and this works unchanged.
set -eu

CONFIG_DIR="${CONFIG_DIR:-/app/build/config}"
ENV_FILE="${CONFIG_DIR}/env.js"

# Keys the frontend understands. Unset keys are omitted entirely so that
# src/config.js falls through to its own defaults.
KEYS="API_URL MAX_PODS REFRESH_INTERVAL_MS HEAT_THRESHOLD SHOW_LABELS SHOW_POD_COUNT SHOW_TOOLTIPS SHOW_FIRE"

escape() {
  printf '%s' "$1" | sed -e 's/\\/\\\\/g' -e 's/"/\\"/g'
}

if ! mkdir -p "$CONFIG_DIR" 2>/dev/null || ! : > "$ENV_FILE" 2>/dev/null; then
  echo "kube-cats: cannot write $ENV_FILE" >&2
  echo "kube-cats: mount a writable volume at $CONFIG_DIR, or set CONFIG_DIR" >&2
  exit 1
fi

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
