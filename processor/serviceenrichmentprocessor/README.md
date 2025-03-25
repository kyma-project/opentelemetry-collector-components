# Service Enrichment Processor

The processor enriches the resource attribute `service_name` if it is not present. The processor follows a default priority
```yaml
    - "k8s.deployment.name",
	- "k8s.daemonset.name",
	- "k8s.statefulset.name",
	- "k8s.job.name",
	- "k8s.pod.name",
```

Additionally one can define additional keys which would be prepended to the default priority. The keys added would follow the
priority in which they are defined.

## Configuration

```yaml
service_name_enrichment:
  custom_labels:
  - "kyma.kubernetes_io_app_name",
  - "kyma.app_name",
```