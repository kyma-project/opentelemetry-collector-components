# Dummy Receiver

| Status      |                            |
|-------------|----------------------------|
| stability   | alpha: metrics             |
| Code Owners | kyma-project/observability |

Dummy Receiver is an OTel Collector receiver that generates dummy telemetry data. At the moment, it only supports metrics. It is useful for testing the OTel Collector pipeline.

## How It Works

It generates dummy metrics and sends them to the OTel Collector pipeline.

For the generated metrics, see [documentation.md](documentation.md).

## Configuration

See the following example configuration:

```yaml
receivers:
  dummy:
    interval: 2s
```
