
# OpenTelemetry Collector Components

## Status

[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/opentelemetry-collector-components)](https://api.reuse.software/info/github.com/kyma-project/opentelemetry-collector-components)

![GitHub tag checks state](https://img.shields.io/github/checks-status/kyma-project/opentelemetry-collector-components/main?label=opentelemetry-collector-components&link=https%3A%2F%2Fgithub.com%2Fkyma-project%2Fopentelemetry-collector-components%2Fcommits%2Fmain)

## Overview

Contains a custom distribution of the [OTel Collector Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib) bundle with additional [OTel Collector](https://github.com/open-telemetry/opentelemetry-collector) components used by the [Kyma Telemetry module](https://github.com/kyma-project/telemetry-manager/tree/main). The additional components are either general and planned to be contributed to the upstream contrib repo, or Kyma-specific.

For actual distribution configuration, see [OTel Collector Docker Image](./otel-collector/).

The additional components are located in the [receiver](./receiver/) folder.

## Prerequisites

TBD: List the requirements to run the project or example.

## Installation

TBD: Explain the steps to install your project. If there are multiple installation options, mention the recommended one and include others in a separate document. Create an ordered list for each installation task.

## Usage

TBD: Explain how to use the project. You can create multiple subsections (H3). Include the instructions or provide links to the related documentation.

## Development

TBD: Add instructions on how to develop the project or example. It must be clear what to do and, for example, how to trigger the tests so that other contributors know how to make their pull requests acceptable. Include the instructions or provide links to related documentation.

## Contributing

See the [Contributing Rules](CONTRIBUTING.md).

## Code of Conduct

See the [Code of Conduct](CODE_OF_CONDUCT.md) document.

## Licensing

See the [license](./LICENSE) file.
