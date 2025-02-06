# OTEL Collector Docker Image

The container image is built using the [OTel builder binary](https://github.com/open-telemetry/opentelemetry-collector/tree/main/cmd/builder) based on [builder-config.yaml](https://github.com/open-telemetry/opentelemetry-collector/blob/main/cmd/otelcorecol/builder-config.yaml).

This custom image has a minimal set of required receivers, processors, and exporters for Kyma.

In addition to OSS components, it contains custom receivers, processors etc., contained in the [OCC repository](https://github.com/kyma-project/opentelemetry-collector-components).

The build is configured for two different modes:

PR build: 
This mode depends on the local version the OTel Collector components and can be used during development to create an image without actually releasing a new version of the OCC repository.

Release build:
This mode relies on a released version of the OCC repository.

## Build locally

1. The build mode defaults to `PR`. To change this, set the **BUILD_MODE** variable to either `PR` or `release`.
1. To build the image locally, run it in the repository root folder:
    ```sh
    docker build -f otel-collector/Dockerfile .
    ```
1. If your build was successful, the Docker command updates its status output to:

       Building {X}s (18/18) FINISHED