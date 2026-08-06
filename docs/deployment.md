# Deployment

## Helm

The chart lives in a separate repo, [j6nca/helm-charts](https://github.com/j6nca/helm-charts):

```bash
helm repo add j6nca https://charts.j6n.dev
helm install kube-cats j6nca/kube-cats --namespace kube-cats --create-namespace
```

Set `ENVIRONMENT=production` on the API so it uses in-cluster credentials
instead of looking for a kubeconfig, and locks CORS to the deployed frontend
origin rather than allowing `*`.

## Images

Built and pushed to GHCR by GitHub Actions on every push to `main` that touches
the matching directory:

| Component | Image |
| --- | --- |
| Backend | `ghcr.io/j6nca/kube-cats-api` |
| Frontend | `ghcr.io/j6nca/kube-cats` |

Both are tagged with the short commit SHA and carry build provenance
attestation. The workflows are path-filtered, so a frontend change does not
rebuild the backend.

## RBAC

The API is read-only and needs `list` on pods and nodes. Fire additionally
needs the metrics API — without it everything still works, but pressure
degrades to the `requests` estimate.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kube-cats
rules:
  - apiGroups: [""]
    resources: ["pods", "nodes"]
    verbs: ["list"]
  # Optional: live CPU/memory usage for the fire sprite.
  - apiGroups: ["metrics.k8s.io"]
    resources: ["pods", "nodes"]
    verbs: ["list"]
```

Bound to the API's service account with a `ClusterRoleBinding`. No `get`,
`watch`, `create` or `update` is required — the scrape only lists, and nothing
in the codebase writes to the cluster.

!!! warning "Check the chart's defaults"
    If you installed the chart before the metrics support landed, its
    `ClusterRole` will not include the `metrics.k8s.io` rule. The symptom is a
    log line on every scrape:

    ```
    metrics-server unavailable for nodes, using requests instead: ...
    ```

    and `"source": "requests"` in the `/pods` response.

## metrics-server

If it is not already installed:

```bash
helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server/
helm install metrics-server metrics-server/metrics-server -n kube-system
```

Verify the API is answering:

```bash
kubectl top nodes
```

If that works and kube-cats still reports `requests`, the RBAC rule above is the
thing to check.

## Verifying a deployment

```bash
# backend reachable and scraping
curl -s https://<api-host>/health

# is pressure real or estimated?
curl -s https://<api-host>/nodes | jq '.[0].pressure'
# {"cpu":0.23,"memory":0.41,"source":"metrics"}
```

`source: "metrics"` means the flame reflects live usage. `source: "requests"`
means it reflects how much has been *reserved*. See
[Pressure](architecture.md#pressure-how-hot-is-hot).

## Configuring the frontend

The web image reads its configuration from the environment **at container
start**, not at build time, so pointing it at a different API is a restart
rather than a rebuild. See [Configuration](configuration.md) for the full list
of keys.
