# Service Enrichment Processor

| Status      |                             |
|-------------|-----------------------------|
| stability   | alpha: metrics,traces, logs |
| Code Owners | kyma-project/observability  |

The processor enriches the [service resource attributes](https://opentelemetry.io/docs/specs/semconv/resource/#service) if they are not present. The processor follows a default priority. Currently it
has been implemented to enrich the service name based on the following attribute keys priority:
```yaml
    - "k8s.deployment.name",
    - "k8s.daemonset.name",
    - "k8s.statefulset.name",
    - "k8s.job.name",
    - "k8s.pod.name",
```

Additionally, you can define additional resource attributes, which are prepended to the default priority. The added resource attributes follow the priority in which they are defined.

## Configuration

```yaml
service_name_enrichment:
  resource_attributes:
  - "kyma.kubernetes_io_app_name",
  - "kyma.app_name",
```