# Local development

## Prerequisites

Versions are pinned in `.tool-versions` (works with [asdf](https://asdf-vm.com/)
or [mise](https://mise.jdx.dev/)):

- Go 1.24
- Node 22.15.1
- A Kubernetes cluster you can reach with `kubectl`

## Proxy the Kubernetes API

In dev mode the backend talks to your cluster through `kubectl proxy` rather
than using in-cluster credentials:

```bash
kubectl proxy --port 8087
```

The backend expects the proxy on `http://127.0.0.1:8087`. Override with
`KUBE_PROXY_URL` if you run it elsewhere:

```bash
KUBE_PROXY_URL=http://127.0.0.1:8001 go run .
```

!!! info "What puts it in dev mode"
    Anything other than `ENVIRONMENT=production`. Dev mode uses your kubeconfig
    via the proxy and allows CORS from any origin; production uses the in-cluster
    service account and locks CORS to the deployed frontend.

## Run the backend

From `api/`:

```bash
go run .
# or
go build -o kube-cat-api && ./kube-cat-api
```

It listens on `:8001` by default (`API_PORT` to change) and exposes:

| Endpoint | Returns |
| --- | --- |
| `GET /health` | `OK` |
| `GET /pods` | Simplified pod list |
| `GET /nodes` | Simplified node list |
| `GET /test` | A fixed string, for connectivity checks |

The first scrape happens immediately at start-up, then every minute
(`SCRAPE_INTERVAL`). Until it completes, the lists are empty.

## Run the frontend

From `web/`:

```bash
npm install
npm start
```

This serves on `:3000` and reads `web/.env`, which already points at
`http://localhost:8001`. For personal overrides use `web/.env.local` — it is
gitignored.

!!! warning "Env var names"
    For local development the variables need the `REACT_APP_` prefix, because
    that is the only thing Create React App exposes to the bundle. Deployed
    containers use the same names *without* the prefix, resolved at runtime.
    [Configuration](configuration.md) explains why.

## Tests

=== "Backend"

    ```bash
    cd api
    go test ./...
    go test -run TestPodPressure ./...   # single test
    go vet ./...
    ```

=== "Frontend"

    ```bash
    cd web
    npm test                             # watch mode
    npm test -- --watchAll=false         # single run
    npm test -- heat.test.js             # single file
    ```

There is no separate lint step; ESLint runs as part of `npm start` and
`npm run build` using the `react-app` config in `package.json`.

## Working without a cluster

`web/src/example/nodes.json` and `pods.json` hold raw `kubectl proxy` output
(full `NodeList`/`PodList`).

!!! danger "These are not the API shape"
    They are the *unsimplified* Kubernetes objects, not what `api/` serves. The
    commented-out `localNodes`/`localPods` requires at the top of `App.js` will
    not render without transforming them first. To develop offline, it is easier
    to serve a stub:

    ```bash
    # returns [] for both endpoints
    npx serve -l 8001 --cors
    ```

## Building the container images

```bash
docker build -t kube-cats-api ./api
docker build -t kube-cats-web ./web
```

The web image regenerates its runtime config at start-up, so you can repoint a
built image without rebuilding:

```bash
docker run -p 3000:3000 -e API_URL=http://localhost:8001 kube-cats-web
```
