# Tiny Pushgateway Helm Chart

A simple Helm chart for deploying the tiny-pushgateway application to Kubernetes.

## Installation

```bash
# Install the chart
helm install tiny-pushgateway ./chart

# Install with custom values
helm install tiny-pushgateway ./chart --set image.tag=v1.0.0

# Install with custom values file
helm install tiny-pushgateway ./chart -f custom-values.yaml
```

## Uninstallation

```bash
helm uninstall tiny-pushgateway
```

## Configuration

The following table lists the configurable parameters of the chart and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Image repository | `endyrocket/tiny-pushgateway` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `image.tag` | Image tag | `latest` |
| `service.type` | Service type | `ClusterIP` |
| `service.port` | Service port | `9091` |
| `resources.limits.cpu` | CPU limit | `100m` |
| `resources.limits.memory` | Memory limit | `128Mi` |
| `resources.requests.cpu` | CPU request | `50m` |
| `resources.requests.memory` | Memory request | `64Mi` |

## Usage

After installation, the tiny-pushgateway will be available at port 9091.

### Push metrics

```bash
kubectl port-forward svc/tiny-pushgateway 9091:9091
curl -X POST http://localhost:9091/push -d 'my_metric 123'
```

### View metrics

```bash
curl http://localhost:9091/metrics
```

### Configure Prometheus to scrape

Add to your Prometheus configuration:

```yaml
scrape_configs:
  - job_name: 'tiny-pushgateway'
    static_configs:
      - targets: ['tiny-pushgateway:9091']
```
