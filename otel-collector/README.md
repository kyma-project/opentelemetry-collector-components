# OTEL Collector Docker Image

The container image is built using the [otel builder binary](https://github.com/open-telemetry/opentelemetry-collector/tree/main/cmd/builder) based on [builder-config.yaml](https://github.com/open-telemetry/opentelemetry-collector/blob/main/cmd/otelcorecol/builder-config.yaml).

This custom image has a minimal set of required receivers, processors and exporters for Kyma.

In addition to OSS components it will contain custom receivers, processors etc contained in the [OCC repository](https://github.com/kyma-project/opentelemetry-collector-components)

The build is configured for two different scenarios:

PR build: 
This mode depends on the local version the opentelemtry-collector-components and can be used during development to create an image without actually releasing a new version of the OCC repository.

Release build:
This mode relies on a released version of the OCC repository.


## Build locally

The build mode defaults to `PR`. To change this set the BUILD_MODE variable to either
`PR` or `release`.

To build the image locally, execute the following command, entering the proper versions taken from the `envs` file:

execute the following command in the repository root folder:

PR-mode:

```
docker build -f otel-collector/Dockerfile --build-arg GOLANG_VERSION=XXX --build-arg='OTEL_VERSION=XXX' --build-arg OTEL_CONTRIB_VERSION=XXX .
```

release-mode:

```
docker build -f otel-collector/Dockerfile --build-arg BUILD_MODE=release --build-arg GOLANG_VERSION=XXX --build-arg OTEL_VERSION=XXX --build-arg OTEL_CONTRIB_VERSION=XXX .
```

