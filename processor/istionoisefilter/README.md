# Istio Noise Filter Processor

| Status      |                             |
|-------------|-----------------------------|
| stability   | alpha: logs, metrics, traces|
| Code Owners | kyma-project/observability  |

The **Istio Noise Filter Processor** is an OpenTelemetry Collector processor that drops noisy technical telemetry signals (traces, logs, and metrics) generated by Istio proxies in a typical Kyma environment. Its goal is to reduce the volume of low-value or redundant telemetry, making it easier to focus on meaningful application signals.

## Usage

Add the processor to your OpenTelemetry Collector pipeline:

```yaml
processors:
  istio_noise_filter:

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [istio_noise_filter]
      exporters: [otlp]
    logs:
      receivers: [otlp]
      processors: [istio_noise_filter]
      exporters: [otlp]
    metrics:
      receivers: [otlp]
      processors: [istio_noise_filter]
      exporters: [otlp]
```

## Development

- Filtering rules are maintained in the `internal/rules` package.
- Unit tests for all rules are provided in the corresponding `*_test.go` files.
