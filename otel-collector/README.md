# OTEL Collector Docker Image

The container image is built using the [otel builder binary](https://github.com/open-telemetry/opentelemetry-collector/tree/main/cmd/builder) based on [builder-config.yaml](https://github.com/open-telemetry/opentelemetry-collector/blob/main/cmd/otelcorecol/builder-config.yaml).

This custom image has a minimal set of required receivers, processors and exporters for Kyma.

## Build locally

To build the image locally, execute the following command, entering the proper versions taken from the `envs` file:
```
docker build -t otel-collector:local --build-arg OTEL_VERSION=XXX --build-arg GOLANG_VERSION=XXX .
```
