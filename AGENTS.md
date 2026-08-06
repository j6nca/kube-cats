# AGENTS.md

Guidance for coding agents working in this repository.

## What this is

A read-only Kubernetes visualizer that renders pods as pixel cats drifting across the screen and nodes as UFOs. Deliberately not a monitoring tool — accuracy and completeness matter less than it being fun and never mutating the cluster.

Two independently built and deployed components:
- `api/` — Go 1.24 backend (module `kube-cat-api`), uses `k8s.io/client-go` and `k8s.io/metrics`; `main.go`, `cache.go`, `handlers.go`, `kube.go`, `cache_test.go`
- `web/` — React 19 frontend on Create React App (`react-scripts`), animation via `motion`/`framer-motion`; source in `web/src/` (`App.js`, `config.js`, `heat.js`, `components/`, `assets/`, `example/`)

Tool versions are pinned in `.tool-versions` (node 22.15.1, go 1.24). Each component has its own `Dockerfile`; deployment is via a Helm chart in the separate `j6nca/helm-charts` repo. User-facing docs are an MkDocs Material site in `docs/`, published to GitHub Pages.

## Commands

Backend (from `api/`):
```
go run .
go test ./...
go test -run TestPodPressure ./...       # single test
go vet ./...
```

Frontend (from `web/`):
```
npm start                                # dev server on :3000
npm run build
npm test                                 # CRA/Jest watch mode
npm test -- --watchAll=false             # single run (use in CI/scripts)
npm test -- heat.test.js                 # single test file
```

Docs (from the repo root, needs `pip install mkdocs-material==9.5.49`):
```
mkdocs serve
mkdocs build --strict                    # what CI runs; broken links fail the build
```

There is no lint script; ESLint runs through `react-scripts` during `start`/`build` using the `react-app` config in `package.json`. `npm run build` is warning-clean — keep it that way, since `CI=true` turns warnings into errors.

## Local development

The backend expects a `kubectl proxy` in dev mode:

```
kubectl proxy --port 8087
```

`api/kube.go` defaults to `http://127.0.0.1:8087` to match; override with `KUBE_PROXY_URL`. The API listens on `$API_PORT` (default `8001`), and `web/.env` already points the frontend there.

## Architecture

**Backend: poll-and-cache, never proxy.** `main.go` builds one client bundle, then a goroutine calls `getKubeInfo` every `SCRAPE_INTERVAL` (default 1m); HTTP handlers only read from the in-memory `Cache` (`cache.go`, guarded by an embedded `sync.RWMutex`). Requests never touch the Kubernetes API, so handler latency is constant and cluster load is independent of traffic. A failed scrape logs and returns, leaving the previous cache intact.

`simplifyPods`/`simplifyNodes` in `cache.go` are the deliberate trust boundary — cluster objects are reduced to `SimplifiedPod`/`SimplifiedNode` before anything is exposed. Add fields there consciously; the whole API surface is those two structs. Resources are published both as Kubernetes quantity strings and as numeric millicores/bytes so the frontend never parses quantity notation.

**Pressure is the fire signal, and its `source` matters.** Each entity carries `pressure: {cpu, memory, source}`. `source` is `metrics` (live, from metrics-server), `requests` (metrics-server unavailable, estimated from resource requests), or `unknown` (no denominator — nothing to divide by). Entities with `unknown` must never render as healthy-looking zeroes; `heat.js` excludes them from catching fire for exactly this reason. Ratios can exceed 1.0 and are not clamped. Denominators are documented in `docs/architecture.md` — pods measure against their own limits, falling back to their node's allocatable.

metrics-server is optional everywhere: a nil metrics client or a failed metrics list degrades that scrape to `requests` rather than dropping it.

**Auth mode is env-driven.** `kube.go` treats `ENVIRONMENT != "production"` as dev (kubeconfig via the local proxy) and production as in-cluster (`rest.InClusterConfig`). `handlers.go` keys CORS off the same variable: `*` in dev, `https://meow.j6n.dev` in production.

Endpoints: `GET /health`, `GET /test`, `GET /pods`, `GET /nodes`.

**Frontend: one fetch, everything else is animation.** `App.js` fetches `/nodes` and `/pods` in parallel on mount and every `REFRESH_INTERVAL_MS`, then renders each entity inside a `motion.div` that translates it from off-screen left to off-screen right on an infinite loop with randomized duration and vertical offset. Nodes and pods are siblings — pods do not render inside their node — so `EntityNode` receives `pods={pods.filter(...)}` purely to display a count. Sprite choice is `cats[podIndex % cats.length]`, i.e. positional, not derived from pod state.

Adding a cat sprite means dropping the gif in `web/src/assets/pods/` and adding an import plus an array entry in `components/EntityPod.js`. The fire overlay (`assets/misc/fire.svg`) is a separate positioned `<img>`, so it composes with every sprite rather than needing per-cat variants.

## Configuration

Three layers, first match wins: `window.__KUBE_CATS_ENV__` (runtime) → `process.env.REACT_APP_*` (build time) → defaults in `config.js`.

The runtime layer exists because CRA inlines `REACT_APP_*` at build time, which freezes a published image to whatever CI had. `web/docker-entrypoint.sh` regenerates `build/env.js` from the container environment on every start, using the same key names **without** the `REACT_APP_` prefix. Unset keys are omitted rather than emitted as empty strings, so they fall through instead of blanking a working default.

When adding a config key you must touch three places: the default in `web/src/config.js`, the `KEYS` list in `web/docker-entrypoint.sh`, and the table in `docs/configuration.md`. Missing the second is the easy mistake — it fails silently, only in containers.

`web/.env` holds local `REACT_APP_*` defaults. It must stay at the `web/` root: CRA does not read a `.env` from `src/`, which is where it used to live and why none of these settings ever took effect.

## Gotchas

`web/src/example/*.json` are raw `kubectl proxy` dumps (full `NodeList`/`PodList`), not the simplified API shape. The commented-out `localNodes`/`localPods` requires in `App.js` will not work against the current components without transformation.

`api/handlers.go` and `api/main.go` are not gofmt-clean (pre-existing mixed tabs/spaces). Don't reformat them wholesale in an unrelated change; the diff noise buries the real edit.

`EntityNode.js` and `App.js` carry intentionally commented-out features (nested pod rendering, the show/hide UI toggle). Leave them; they are not dead code the author forgot about.

## CI/CD

Three path-filtered workflows in `.github/workflows/`, so each change rebuilds only what it touches:

| Workflow | Triggers on | Does |
| --- | --- | --- |
| `api.yaml` | `api/**` | Builds and pushes `ghcr.io/<repo>-api` |
| `web.yaml` | `web/**` | Builds and pushes `ghcr.io/<repo>` |
| `docs.yaml` | `docs/**`, `mkdocs.yml` | `mkdocs build --strict` → GitHub Pages |

The image builds include provenance attestation. There is no test or lint job — tests run locally only.
