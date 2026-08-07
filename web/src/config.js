// Configuration is resolved in three layers, first match wins:
//
//   1. window.__KUBE_CATS_ENV__  - runtime, written into build/config/env.js by
//      docker-entrypoint.sh from the container's environment. This is what
//      makes a single published image configurable per-deployment.
//   2. process.env.REACT_APP_*   - build time, baked in by Create React App.
//      Useful for `npm start` via web/.env; useless for a published image.
//   3. the defaults below.
//
// Note the REACT_APP_ prefix is mandatory for layer 2: CRA strips any other
// variable, which is why the previous unprefixed reads always fell through.

const runtime = (typeof window !== 'undefined' && window.__KUBE_CATS_ENV__) || {};

const raw = (key) => {
  const value = runtime[key] ?? process.env[`REACT_APP_${key}`];
  return value === undefined || value === null || value === '' ? undefined : value;
};

const str = (key, fallback) => raw(key) ?? fallback;

const bool = (key, fallback) => {
  const value = raw(key);
  if (value === undefined) return fallback;
  return String(value).toLowerCase() === 'true';
};

const num = (key, fallback) => {
  const value = Number(raw(key));
  return Number.isFinite(value) ? value : fallback;
};

const config = {
  api_url: str('API_URL', 'https://kube-cats-api.j6n.dev'),
  environment: process.env.NODE_ENV || 'development',

  // Rendering every pod in a large cluster is a good way to melt a laptop.
  max_pods: num('MAX_PODS', 200),
  refresh_interval_ms: num('REFRESH_INTERVAL_MS', 5 * 60 * 1000),

  // Pods and nodes at or above this share of their cpu/memory catch fire.
  heat_threshold: num('HEAT_THRESHOLD', 0.8),

  show_labels: bool('SHOW_LABELS', true),
  show_pod_count: bool('SHOW_POD_COUNT', true),
  show_tooltips: bool('SHOW_TOOLTIPS', false),
  show_fire: bool('SHOW_FIRE', true),
};

export default config;
