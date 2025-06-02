# Istio Enrichment Processor

| Status      |                            |
|-------------|----------------------------|
| stability   | alpha: logs                |
| Code Owners | kyma-project/observability |


The processor enriches the Istio access log attributes with the following log attributes:

- `client.address`
- `client.port`
- `network.protocol.name`
- `network.protocol.version`

Additionally, the processor sets the log severity attributes and instrumentation scope attributes to the following values:

- `severity.text` to `INFO`
- `severity.number` to `9`
- `scope.name` to `io.kyma-project.telemetry/istio`
- `scope.version` to `<Kyma Telemetry Module version>`

## Configuration

```yaml
processors:
    istio_enrichment:
        scope_version: 1.41.0

service:
  pipelines:
    logs:
      receivers:
        - otlp
      processors:
        - istio_enrichment
      exporters:
        - otlp
```