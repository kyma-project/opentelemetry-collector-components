# Kyma Stats Receiver


| Status      |                            |
|-------------|----------------------------|
| stability   | alpha: metrics             |
| Code Owners | kyma-project/observability |

The Kyma Stats Receiver pulls Kyma modules from the API server, creates status metrics, and sends them down the metric pipeline for further processing.

## Metrics

For details about the metrics produced by the Kyma Stats Receiver, see [metadata.yaml](./metadata.yaml) and [documentation.md](./documentation.md)

## Configuration

The following settings are required:

- `auth_type` (default = `serviceAccount`): Specifies the authentication method for accessing the Kubernetes API server.
   Options include `none` (no authentication), `serviceAccount` (uses the default service account token assigned to the Pod), or `kubeConfig` (uses credentials from `~/.kube/config`).
- `module_groups`: A list of API groups to be used for Kyma module resource discovery. The receiver then discovers all resources within these groups, using their preferred versions, except those specified in `excluded_resources`. For each group-version-resource, the first resource instance is selected for status metric generation.

The following settings are optional:

- `collection_interval` (default = `60s`): The Kyma Stats Receiver monitors Kyma custom resources using the Kubernetes API. It emits the collected metrics only once per collection interval. The `collection_interval` setting determines how frequently these metrics are emitted.
- `excluded_resources`: A list of API resource names to be excluded from status metrics generation.
- `metrics`: Enables or disables specific metrics.
- `resource_attributes`: Enables or disables resource attributes.

Example:

```yaml
  kymastats:
    auth_type: seviceAccount
    collection_interval: 30s
    excluded_resources:
    - kymas
    - moduletemplates
    metrics:
      kyma.module.status.state:
        enabled: true
      kyma.module.status.conditions:
        enabled: true
    module_groups:
    - operator.kyma-project.io
    resource_attributes:
      k8s.namespace.name:
        enabled: true
      kyma.module.name:
        enabled: true
```

For the full list of settings exposed for the Kyma Stats Receiver, see the [config.go](./config.go) file.
For detailed sample configurations , see the [config.yaml](./testdata/expected_config.yaml) file.
