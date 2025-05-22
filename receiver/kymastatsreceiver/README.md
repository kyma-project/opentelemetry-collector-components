# Kyma Stats Receiver


| Status      |                            |
|-------------|----------------------------|
| stability   | alpha: metrics             |
| Code Owners | kyma-project/observability |

The Kyma Stats Receiver pulls Kyma resources from the API server, creates status metrics, and sends them down the metric pipeline for further processing.

## Metrics

For details about the metrics produced by the Kyma Stats Receiver, see [metadata.yaml](./metadata.yaml) and [documentation.md](./documentation.md)

## Configuration

The following settings are required:

- `auth_type` (default = `serviceAccount`): Specifies the authentication method for accessing the Kubernetes API server.
   Options include `none` (no authentication), `serviceAccount` (uses the default service account token assigned to the Pod), or `kubeConfig` (uses credentials from `~/.kube/config`).
- `k8s_leader_elector`: References the k8s leader elector extension.
- `resources`: A list of API group-version-resources of Kyma resources. Status metrics are generated for each group-version-resource.

The following settings are optional:

- `collection_interval` (default = `60s`): The Kyma Stats Receiver monitors Kyma custom resources using the Kubernetes API. It emits the collected metrics only once per collection interval. The `collection_interval` setting determines how frequently these metrics are emitted.
- `metrics`: Enables or disables specific metrics.
- `resource_attributes`: Enables or disables resource attributes.

Example:

```yaml
  kymastats:
    auth_type: seviceAccount
    collection_interval: 30s
    metrics:
      kyma.resource.status.state:
        enabled: true
      kyma.resource.status.conditions:
        enabled: true
    resources:
    - group: operator.kyma-project.io
      version: v1alpha1
      resource: telemetries
    resource_attributes:
      k8s.namespace.name:
        enabled: true
      k8s.resource.name:
        enabled: true
```

For the full list of settings exposed for the Kyma Stats Receiver, see the [config.go](./config.go) file.
For detailed sample configurations , see the [config.yaml](./testdata/config.yaml) file.
