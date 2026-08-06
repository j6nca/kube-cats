# kube-cats

![kube-cats uptime](https://status.j6n.dev/api/v1/endpoints/projects_kube-cats/uptimes/7d/badge.svg)
![kube-cats health](https://status.j6n.dev/api/v1/endpoints/projects_kube-cats/health/badge.svg)

A fun kubernetes visualizer with pixel cats! 🐱

📖 **[Documentation](https://j6nca.github.io/kube-cats/)**

> [!NOTE] 
> This project is meant just for fun and should not be used in place of any monitoring/metrics tooling.

## Goals

- [x] display various components of the cluster in a read-only, fun, form factor
- [x] semi-randomized asset usage (different combinations of cats)
- [x] special cases (based off pod health, etc.) — pods and nodes over their cpu/memory threshold catch fire
- [ ] add support for Gateway resources
- [x] fix dynamic configuration — see [Configuration](https://j6nca.github.io/kube-cats/configuration/)

## Installation

> [!NOTE] 
> Not yet ready.

You can install the chart from my [charts repo](https://github.com/j6nca/helm-charts), or make your own.

```
helm repo add j6nca https://charts.j6n.dev
helm install kube-cats j6nca/kube-cats --namespace kube-cats
```

## Running Locally

Proxy your Kubernetes API locally:

```
kubectl proxy --port 8087
```

Run backend (from `api/`):

```
go run .
```

Run frontend (from `web/`):

```
npm install
npm run start
```

Local defaults live in `web/.env`; the frontend points at `http://localhost:8001` out of the box.

> [!TIP]
> Install [metrics-server](https://github.com/kubernetes-sigs/metrics-server) to get real cpu/memory
> usage behind the fire sprite. Without it, kube-cats falls back to resource requests and says so in
> the `pressure.source` field.

## Inspired By
- [kube-ops-view by hjacobs](https://codeberg.org/hjacobs/kube-ops-view)
- [flying-k8s by milagrofrost](https://github.com/milagrofrost/Flying-K8s)
