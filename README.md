# tiny-pushgateway

A minimal, ultra-lightweight, high-performance Prometheus pushgateway with auto-clearing buffers, designed for ephemeral workloads and short-lived jobs.

---

## Why?

Short-lived jobs (batch processes, cron jobs, CI/CD tasks) terminate before Prometheus can scrape them directly. These workloads require an intermediate gateway to push metrics to, which Prometheus can then scrape.

### The Problem with Prometheus Pushgateway

While the [official Prometheus Pushgateway](https://github.com/prometheus/pushgateway) provides this functionality, it has a critical flaw: **it persists metrics indefinitely**. Once pushed, metrics remain in the gateway until explicitly deleted. This causes:

- **Stale metrics** accumulate and get scraped repeatedly
- **Time series pollution** — Prometheus records the same value multiple times with different timestamps
- **Incorrect aggregations** — stale data skews queries and alerts
- **Manual cleanup** required (DELETE requests or restarts)

This design violates Prometheus's pull model philosophy and is [explicitly documented](https://prometheus.io/docs/practices/pushing/) as problematic for short-lived jobs.

### How tiny-pushgateway Solves This

A **stateless relay** with auto-clearing buffers — metrics are exposed **once** and immediately cleared after Prometheus scrapes them. No stale data, no time series pollution, no manual cleanup.

Perfect for:
- Kubernetes Jobs and CronJobs
- CI/CD pipelines
- Serverless functions
- Lambda / Cloud Run tasks
- Any ephemeral workload that can't be scraped directly

---

## Repository layout
```
tiny-pushgateway/
├── main.go # HTTP server (/push, /metrics)
├── main_test.go # unit & end-to-end tests
├── Dockerfile # multi-stage build → ~7 MB distroless image
├── go.mod # Go 1.22 module definition
└── README.md # you are here
```
---

## Features

* **Auto-clearing buffer**  
  • `POST /push` — append text-format samples to an in-memory buffer  
  • `GET  /metrics` — expose **and clear** the buffer (one-time scrape)
* **Format validation** — rejects malformed Prometheus exposition format
* **Retry resilience** — buffers metrics through up to 3 failed scrapes
* **Zero dependencies** beyond the Go standard library
* **Tiny footprint**   
  • Small attack surface
  • Docker image ~7 MB
* **Ultra fast** — handles 6k+ requests/sec

---

## Quick start

### Run locally
```bash
go run .
```
### Push & scrape
```bash
echo 'demo_metric{job="test"} 1' | curl -X POST --data-binary @- http://localhost:9091/push
curl http://localhost:9091/metrics
```
### Docker
```bash
docker build -t tiny-pushgateway .
docker run -p 9091:9091 tiny-pushgateway
```

### Kubernetes (Helm)
```bash
# Install the chart
helm install tiny-pushgateway ./chart

# Get the pod name and port-forward
export POD_NAME=$(kubectl get pods -l "app.kubernetes.io/name=tiny-pushgateway,app.kubernetes.io/instance=tiny-pushgateway" -o jsonpath="{.items[0].metadata.name}")
kubectl port-forward $POD_NAME 9091:9091

# Push metrics
curl -X POST http://localhost:9091/push -d 'job_metric{status="success"} 1'

# Verify
curl http://localhost:9091/metrics
```

**Configure Prometheus to scrape:**
```yaml
scrape_configs:
  - job_name: 'tiny-pushgateway'
    scrape_interval: 15s
    static_configs:
      - targets: ['tiny-pushgateway:9091']
```

**Customize deployment:**
```bash
# Set custom image and resources
helm install tiny-pushgateway ./chart \
  --set image.repository=myregistry/tiny-pushgateway \
  --set image.tag=v1.0.0 \
  --set resources.limits.memory=256Mi

# Use custom values file
helm install tiny-pushgateway ./chart -f custom-values.yaml
```

See [`chart/README.md`](chart/README.md) for more configuration options.

---

## Performance

Benchmark results with **1 CPU** and **1 GB memory**:

```
  Total:	15.0085 secs
  Slowest:	0.1052 secs
  Fastest:	0.0029 secs
  Average:	0.0152 secs
  Requests/sec:	**6562.9631**

Response time histogram:
  0.003 [1]	|
  0.013 [46452]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.023 [43450]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.034 [8244]	|■■■■■■■
  0.044 [159]	|
  0.054 [88]	|
  0.064 [14]	|
  0.074 [36]	|
  0.085 [29]	|
  0.095 [21]	|
  0.105 [6]	|

```
