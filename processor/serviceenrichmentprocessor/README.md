# Service Enrichment Processor

| Status      |                             |
|-------------|-----------------------------|
| stability   | alpha: metrics,traces, logs |
| Code Owners | kyma-project/observability  |

The processor enriches the [service](https://opentelemetry.io/docs/specs/semconv/resource/#service) resource attribute if it is not present. The processor follows a default priority. Currently it
has been implemented to enrich the service name based on the following keys:
```yaml
    - "k8s.deployment.name",
    - "k8s.daemonset.name",
    - "k8s.statefulset.name",
    - "k8s.job.name",
    - "k8s.pod.name",
```

Additionally, you can define additional keys, which are prepended to the default priority. The added keys follow the priority in which they are defined.

## Configuration

```yaml
service_name_enrichment:
  custom_labels:
  - "kyma.kubernetes_io_app_name",
  - "kyma.app_name",
```