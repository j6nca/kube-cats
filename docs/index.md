# kube-cats 🐱

A fun Kubernetes visualizer that renders your cluster as pixel cats drifting across the screen.
Pods are cats, nodes are UFOs, and anything running hot catches fire.

!!! note
    This project is meant just for fun and should not be used in place of any
    monitoring or metrics tooling. It is strictly read-only — it never mutates
    the cluster.

## What it shows

| Cluster thing | On screen |
| --- | --- |
| Pod | A pixel cat, sprite chosen by position |
| Node | A UFO, labelled with its pod count |
| Pod or node over its CPU/memory threshold | An animated flame on the sprite |

Hovering a sprite shows its name, and with [`SHOW_TOOLTIPS`](configuration.md#display)
enabled you get namespace, phase, restart count and live utilisation too.

## How it works

Two small components:

- **`api/`** — a Go service that polls the Kubernetes API on a timer and serves
  a heavily trimmed snapshot from memory. Requests never touch the cluster.
- **`web/`** — a React single-page app that fetches that snapshot and animates it.

See [Architecture](architecture.md) for the details, including how the
"running hot" signal is derived.

## Quick start

Point a local backend at your cluster and open the frontend:

```bash
# 1. proxy the Kubernetes API
kubectl proxy --port 8087

# 2. backend, in api/
go run .

# 3. frontend, in web/
npm install && npm start
```

Full instructions in [Local development](local-development.md).

## Getting fire to work

The flame needs to know how busy things are. That comes from
[metrics-server](https://github.com/kubernetes-sigs/metrics-server):

- **With metrics-server** — real live CPU and memory usage.
- **Without it** — the API falls back to resource *requests* versus allocatable
  capacity, and labels the data `requests` so you know it is an estimate.

Nothing breaks without metrics-server; the numbers just mean something
different. See [Pressure](architecture.md#pressure-how-hot-is-hot) and the
[RBAC](deployment.md#rbac) it needs.

## Inspired by

- [kube-ops-view by hjacobs](https://codeberg.org/hjacobs/kube-ops-view)
- [flying-k8s by milagrofrost](https://github.com/milagrofrost/Flying-K8s)
