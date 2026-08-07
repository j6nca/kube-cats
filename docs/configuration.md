# Configuration

## Backend

Plain environment variables read at start-up.

| Variable | Default | Meaning |
| --- | --- | --- |
| `API_PORT` | `8001` | Port to listen on |
| `ENVIRONMENT` | *(unset)* | `production` uses in-cluster auth and locks CORS; anything else is dev mode |
| `KUBECONFIG` | `$HOME/.kube/config` | Kubeconfig path, dev mode only |
| `KUBE_PROXY_URL` | `http://127.0.0.1:8087` | Where `kubectl proxy` is listening, dev mode only |
| `SCRAPE_INTERVAL` | `1m` | How often to poll the cluster. Go duration, e.g. `30s`, `5m` |
| `CORS_ALLOW_ORIGINS` | *(see below)* | Comma-separated browser origins allowed to read the API |

An invalid `SCRAPE_INTERVAL` logs a warning and falls back to `1m` rather than
failing to start.

### CORS

A matching request `Origin` is echoed back, so several hostnames can share one
deployment — a single fixed header can only ever satisfy one of them:

```bash
CORS_ALLOW_ORIGINS=https://meow.j6n.dev,http://kube-cats.j6n.internal
```

`*` allows everything. An origin that is not listed gets no
`Access-Control-Allow-Origin` header at all, so the browser blocks the read.
Requests without an `Origin` (curl, uptime checks) are unaffected — CORS is a
browser mechanism.

When unset the behaviour is the historical one: `*` in dev, and
`https://meow.j6n.dev` alone in production.

!!! tip "Developing against a deployed API"
    Running `npm start` against a remote API fails CORS unless `localhost` is
    in the list, because the page origin is `http://localhost:3000`. Add it to
    `CORS_ALLOW_ORIGINS` on that deployment, or just run the backend locally —
    in dev mode it allows every origin.

## Frontend

The frontend resolves each setting through three layers, first match wins:

1. **Runtime** — `window.__KUBE_CATS_ENV__`, written into `build/config/env.js` when the
   container starts. Keys are used **without** a prefix.
2. **Build time** — `process.env.REACT_APP_*`, baked in by Create React App.
   Set these in `web/.env` for local development.
3. **The built-in defaults** below.

| Key | Default | Meaning |
| --- | --- | --- |
| `API_URL` | `https://kube-cats-api.j6n.dev` | Base URL of the backend |
| `MAX_PODS` | `200` | Cap on pods rendered, to keep the browser alive |
| `REFRESH_INTERVAL_MS` | `300000` | How often the page refetches (5 minutes) |
| `HEAT_THRESHOLD` | `0.8` | Utilisation ratio at which a sprite catches fire |

### Display

| Key | Default | Meaning |
| --- | --- | --- |
| `SHOW_FIRE` | `true` | Master switch for the flame overlay |
| `SHOW_LABELS` | `true` | Name label under each sprite |
| `SHOW_POD_COUNT` | `true` | Pod count next to each node |
| `SHOW_TOOLTIPS` | `false` | Detailed hover tooltip instead of just the name |

Booleans are compared case-insensitively against `true`; anything else is false.
Numbers that do not parse fall back to the default.

!!! tip "Setting the threshold"
    `HEAT_THRESHOLD` is a ratio, not a percentage — `0.8` means 80%. A sprite
    burns when **either** CPU or memory crosses it. If everything is on fire,
    check whether your pressure data is coming from `requests` rather than
    `metrics`; see [Pressure](architecture.md#pressure-how-hot-is-hot).

## Why there are two layers

Create React App inlines `REACT_APP_*` variables into the bundle **at build
time**. For a published container image that is useless: the values are frozen
to whatever the CI runner had, so the same image can never be pointed at a
different API.

So the image ships an entrypoint that writes the environment into a small
script the page loads before the bundle:

```js
// build/config/env.js, regenerated on every container start
window.__KUBE_CATS_ENV__ = {
  "API_URL": "https://kube-cats-api.example.com",
  "HEAT_THRESHOLD": "0.65",
};
```

Unset variables are omitted entirely rather than written as empty strings, so
they fall through to the next layer instead of blanking a working default.
Reconfiguring is a pod restart, not a rebuild.

??? note "Previous behaviour"
    Before this, `config.js` read unprefixed names like `process.env.API_URL`
    and the `.env` file lived in `web/src/`. CRA only exposes `REACT_APP_*`, and
    only reads `.env` from the project root, so **every** value silently fell
    through to its hard-coded default. If you had set these and seen no effect,
    that is why. The keys `MAX_PODS_PER_NODE` and the unused `SHOW_*` entries
    were replaced by the table above; `MAX_PODS_PER_NODE` is now `MAX_PODS`,
    since pods are not drawn per-node.

## Example

=== "Helm values"

    ```yaml
    web:
      env:
        API_URL: https://kube-cats-api.example.com
        HEAT_THRESHOLD: "0.7"
        SHOW_TOOLTIPS: "true"

    api:
      env:
        ENVIRONMENT: production
        SCRAPE_INTERVAL: 30s
    ```

=== "docker run"

    ```bash
    docker run -p 3000:3000 \
      -e API_URL=https://kube-cats-api.example.com \
      -e HEAT_THRESHOLD=0.7 \
      -e SHOW_TOOLTIPS=true \
      ghcr.io/j6nca/kube-cats
    ```

=== "web/.env (local)"

    ```bash
    REACT_APP_API_URL=http://localhost:8001
    REACT_APP_HEAT_THRESHOLD=0.7
    REACT_APP_SHOW_TOOLTIPS=true
    ```
