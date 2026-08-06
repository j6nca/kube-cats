# Architecture

Two components that know almost nothing about each other: a Go API that
summarises the cluster, and a React page that animates the summary.

```
kubectl proxy / in-cluster SA
            │
            ▼
   ┌─────────────────┐   every SCRAPE_INTERVAL
   │  api/ (Go)      │◄──────────────────────────┐
   │                 │                           │
   │  scrape ────────┼──► simplify ──► Cache ────┘
   │                 │                  │
   │  /pods /nodes ◄─────────────────────┘
   └─────────────────┘
            │  JSON, every REFRESH_INTERVAL_MS
            ▼
   ┌─────────────────┐
   │  web/ (React)   │  motion.div per entity
   └─────────────────┘
```

## Backend: poll and cache, never proxy

`main.go` builds one client set, then a goroutine calls `getKubeInfo` on a
ticker. The HTTP handlers only ever read an in-memory `Cache` guarded by an
`sync.RWMutex` — a request never reaches the Kubernetes API.

Two consequences worth knowing:

- Handler latency is constant, and cluster load is independent of how many
  people have the page open.
- Data is up to one scrape interval stale. That is fine for cats.

A failed scrape logs and returns, leaving the previous cache intact, so a brief
API blip shows stale sprites rather than an empty screen.

## The trust boundary

`simplifyPods` and `simplifyNodes` in `cache.go` reduce full Kubernetes objects
to two small structs before anything leaves the process. This is deliberate:
pod specs carry env vars, annotations and image references that have no business
being served to a browser.

The entire public API surface is `SimplifiedPod` and `SimplifiedNode`. Adding a
field there is a conscious decision to publish it.

```json
{
  "name": "api-7d9f",
  "namespace": "default",
  "node": "k-node-1",
  "phase": "Running",
  "ready": true,
  "restarts": 0,
  "resources": {
    "cpuRequests": "100m",
    "cpuRequestsMilli": 100,
    "cpuUsageMilli": 750,
    "memoryUsageBytes": 134217728
  },
  "pressure": { "cpu": 0.75, "memory": 0.12, "source": "metrics" }
}
```

Resources are published twice: the original Kubernetes quantity strings
(`100m`, `128Mi`) and numeric equivalents in millicores and bytes, so the
frontend does not have to parse quantity notation.

## Pressure: how hot is hot

`pressure` is a ratio per resource plus the `source` it came from. The frontend
compares it against `HEAT_THRESHOLD` to decide whether to draw a flame.

`source` takes one of three values:

`metrics`

:   Live usage from metrics-server. The real thing.

`requests`

:   metrics-server was unavailable, so the ratio is estimated from resource
    *requests*. This means "heavily requested", not "actually busy" — a pod
    idling with large requests will look hot, and a pod with no requests at all
    will never look hot.

`unknown`

:   There was no sensible denominator — no limits, no allocatable capacity, e.g.
    an unscheduled pod. **These never catch fire.** Reporting 0% would make them
    indistinguishable from genuinely idle ones, so the frontend skips them
    entirely rather than implying health it cannot verify.

### What each ratio divides by

| Entity | With metrics-server | Fallback |
| --- | --- | --- |
| Pod **with** limits | usage ÷ its own limits | requests ÷ its own limits |
| Pod **without** limits | usage ÷ its node's allocatable | requests ÷ its node's allocatable |
| Node | usage ÷ allocatable | Σ requests of its pods ÷ allocatable |

Measuring a limitless pod against its node is what keeps a single pod eating a
whole node visible, instead of silently unmeasurable.

Ratios can exceed `1.0` — a pod above its requests, or a node overcommitted.
Nothing clamps them, so treat the value as "≥ threshold means hot" rather than
as a percentage to render directly.

Metrics degradation is per-scrape, not per-process: if metrics-server goes away
mid-life the next scrape quietly switches to `requests` and switches back when
it returns.

## Frontend: one fetch, everything else is animation

`App.js` fetches `/nodes` and `/pods` in parallel on mount and on a timer, then
renders each entity inside a `motion.div` that translates it from off-screen
left to off-screen right, on an infinite loop with a randomised duration and
vertical offset.

Two things that surprise people reading it:

- **Nodes and pods are siblings, not nested.** Pods do not render inside their
  node. `EntityNode` receives `pods={pods.filter(...)}` purely to display a
  count. There is commented-out code in `EntityNode.js` for drawing them nested.
- **Sprite choice is positional**, `cats[podIndex % cats.length]` — it is not
  derived from pod state and is not stable across refreshes if the pod list
  reorders.

The flame is an overlay `<img>` positioned over the sprite, not a different
sprite, so it composes with all nine cats and the UFO.

### Adding a cat

Drop the gif in `web/src/assets/pods/`, then add an import and an array entry
in `components/EntityPod.js`. The modulo picks it up automatically.

## The fire sprite

`web/src/assets/misc/fire.svg` is a three-frame pixel flame. Each frame is a
group shown for a third of the cycle, with the animation delays offset by a
third so exactly one is ever visible — a sprite sheet expressed as CSS rather
than a gif, which keeps it a few hundred bytes and lets it respect
`prefers-reduced-motion` (the animation stops and frame one stays up).
