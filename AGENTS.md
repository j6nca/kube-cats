# AGENTS.md

Guidance for coding agents working in this repository.

## What this is

A read-only Kubernetes visualizer that renders pods as pixel cats drifting across the screen and nodes as UFOs. Deliberately not a monitoring tool — accuracy and completeness matter less than it being fun and never mutating the cluster.

Two independently built and deployed components:
- `api/` — Go 1.24 backend (module `kube-cat-api`), uses `k8s.io/client-go`; `main.go`, `cache.go`, `handlers.go`, `kube.go`
- `web/` — React 19 frontend on Create React App (`react-scripts`), animation via `motion`/`framer-motion`; source in `web/src/` (`App.js`, `components/`, `assets/`, `example/`, `config.js`)

Tool versions are pinned in `.tool-versions` (node 22.15.1, go 1.24). Each component has its own `Dockerfile`; deployment is via a Helm chart in the separate `j6nca/helm-charts` repo.

## Commands

Backend (from `api/`):
```
go build -o kube-cat-api && ./kube-cat-api   # or: go install kube-cat-api && kube-cat-api
go vet ./...
```

Frontend (from `web/`):
```
npm start                                # dev server on :3000
npm run build
npm test                                 # CRA/Jest watch mode
npm test -- --watchAll=false             # single run (use in CI/scripts)
npm test -- App.test.js                  # single test file
```

There is no lint script; ESLint runs through `react-scripts` during `start`/`build` using the `react-app` config in `package.json`.

## Local development

The backend expects a `kubectl proxy` in dev mode:

```
kubectl proxy --port 8000
```

Note the port conflict in the docs: `README.md` says `--port 8087`, but `api/kube.go` hard-codes `http://127.0.0.1:8000` as the proxy address. Use 8000, or change `kube.go`.

The API listens on `$API_PORT` (default `8001`). The frontend's default `api_url` in `web/src/config.js` points at the deployed production API — to hit a local backend you must swap the commented-out `http://localhost:8001` line back in.

## Architecture

**Backend: poll-and-cache, never proxy.** `main.go` builds one clientset, then a goroutine calls `getKubeInfo` every minute; HTTP handlers only read from the in-memory `Cache` (`cache.go`, guarded by an embedded `sync.RWMutex`). Requests never touch the Kubernetes API, so handler latency is constant and cluster load is independent of traffic. A failed scrape logs and returns, leaving the previous cache intact.

`simplifyPods`/`simplifyNodes` in `cache.go` are the deliberate trust boundary — the cluster objects are reduced to `SimplifiedPod`/`SimplifiedNode` (name, namespace, node, summed CPU/memory requests+limits; node capacity/allocatable) before anything is exposed. Add fields there consciously; the whole API surface is those two structs.

**Auth mode is env-driven.** `kube.go` treats `ENVIRONMENT != "production"` as dev (kubeconfig via the local proxy) and production as in-cluster (`rest.InClusterConfig`). `handlers.go` keys CORS off the same variable: `*` in dev, `https://meow.j6n.dev` in production.

Endpoints: `GET /health`, `GET /pods`, `GET /nodes`.

**Frontend: one fetch, everything else is animation.** `App.js` fetches `/nodes` and `/pods` in parallel on mount and every 5 minutes, then renders each entity inside a `motion.div` that translates it from off-screen left to off-screen right on an infinite loop with randomized duration and vertical offset. Nodes and pods are siblings — pods do not render inside their node — so `EntityNode` receives `pods={pods.filter(...)}` purely to display a count. Sprite choice is `cats[podIndex % cats.length]`, i.e. positional, not derived from pod state.

Adding a cat sprite means dropping the gif in `web/src/assets/pods/` and adding an import plus an array entry in `components/EntityPod.js`.

## Configuration gotchas

`web/src/config.js` reads `process.env.API_URL`, `MAX_PODS_PER_NODE`, etc. — none of which CRA injects, since CRA only exposes `REACT_APP_*`-prefixed variables, and only from a `.env` at the project root. The existing `.env` sits at `web/src/.env` (wrong directory) and mixes prefixed and unprefixed keys. `react-dotenv` and the empty `web/public/env.js` are the beginnings of a runtime-config approach that isn't wired up. In practice every config value falls back to its hard-coded default. This is the open "fix dynamic configuration" item in `README.md` — don't assume env vars take effect until it's fixed.

Only `max_pods` is actually consulted (it caps rendering in `App.js`); `show_labels`, `show_pod_count`, and `show_tooltips` are read nowhere.

`web/src/example/*.json` are raw `kubectl proxy` dumps (full `NodeList`/`PodList`), not the simplified API shape. The commented-out `localNodes`/`localPods` requires in `App.js` will not work against the current components without transformation.

`web/src/App.test.js` is still the untouched CRA template test and fails against the real `App`.

## CI/CD

`.github/workflows/api.yaml` and `web.yaml` are path-filtered on `api/**` and `web/**` respectively — a change under one directory builds and pushes only that image to `ghcr.io` (`<repo>-api` and `<repo>`), with build provenance attestation. There is no test or lint job.
