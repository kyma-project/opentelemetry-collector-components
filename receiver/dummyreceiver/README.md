# Dummy Receiver

| Status      |                            |
|-------------|----------------------------|
| stability   | alpha: metrics             |
| Code Owners | kyma-project/observability |

Dummy Metrics Receiver is an OTel Collector receiver that generates dummy telemetry data. At the moment it only supports metrics. It is useful when you want to test
the OTel Collector pipeline.

## How It Works

It generates dummy metrics and sends them to the OTel Collector pipeline.

## Configuration

Below is an example of the configuration:

```yaml
receivers:
  dummy:
    interval: 2s
```
