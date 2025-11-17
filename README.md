# kube-cats

![kube-cats uptime](https://status.j6n.dev/api/v1/endpoints/projects_kube-cats/uptimes/7d/badge.svg)
![kube-cats health](https://status.j6n.dev/api/v1/endpoints/projects_kube-cats/health/badge.svg)

A fun kubernetes visualizer with pixel cats! 🐱

> [!NOTE] 
> This project is meant just for fun and should not be used in place of any monitoring/metrics tooling.

## Goals

- [x] display various components of the cluster in a read-only, fun, form factor
- [x] semi-randomized asset usage (different combinations of cats)
- [ ] special cases (based off pod health, etc.)

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

Run backend:

```
go install kube-cat-api
kube-cat-api
```

Run frontend:

```
npm run start
```
## Inspired By
- [kube-ops-view by hjacobs](https://codeberg.org/hjacobs/kube-ops-view)
- [flying-k8s by milagrofrost](https://github.com/milagrofrost/Flying-K8s)
