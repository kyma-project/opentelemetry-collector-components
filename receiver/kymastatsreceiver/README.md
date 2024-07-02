# Kyma Stats Receiver


| Status      |                            |
|-------------|----------------------------|
| stability   | alpha: metrics             |
| Code Owners | kyma-project/observability |

The Kyma Stats Receiver pulls kyma modules from the API server, create status metrics and sends it down the metric pipeline for further processing.

## Metrics

Details about the metrics produced by this receiver can be found in [metadata.yaml](./metadata.yaml) with further documentation in [documentation.md](./documentation.md)

## Configuration

The following settings are required:

- `auth_type` (default = `serviceAccount`): Specifies the authentication method for accessing the K8s API server. 
   Options include `none` (no authentication), `serviceAccount` (uses the default service account token assigned to the pod), or `kubeConfig` (uses credentials from `~/.kube/config`).

The following settings are optional:

- `collection_interval` (default = `60s`): This receiver monitors Kyma custom resources using the K8s API, it emits the collected metrics only once per collection interval. The `collection_interval` setting determines how frequently these metrics are emitted.
- `metrics`: Enables or disables specific metrics.
- `resource_attributes`: Enables or disables resource attributes.

Example:

```yaml
  kymastatsreceiver:
    auth_type: kubeConfig
    collection_interval: 30s
    metrics:
      kyma.module.status.condition:
        enabled: false
    resource_attributes:
      k8s.namespace.name:
        enabled: false
```

The full list of settings exposed for this receiver are documented [here](./config.go)
with detailed sample configurations [here](./testdata/config.yaml).