
FROM --platform=$BUILDPLATFORM golang:1.25.1-alpine3.22 AS build
ARG TARGETOS
ARG TARGETARCH

ARG OTEL_VERSION
ARG OTEL_CONTRIB_VERSION=${OTEL_VERSION}

RUN apk --update add ca-certificates git

ADD receiver /receiver/
ADD internal /internal/
ADD processor /processor/
WORKDIR /app
COPY otel-collector/builder-config.yaml builder-config.yaml

ENV OTEL_VERSION=${OTEL_VERSION}
ENV OTEL_CONTRIB_VERSION=${OTEL_CONTRIB_VERSION}
RUN sed -i s/OTEL_VERSION/${OTEL_VERSION}/g builder-config.yaml
RUN sed -i s/OTEL_CONTRIB_VERSION/${OTEL_CONTRIB_VERSION}/g builder-config.yaml

RUN go install go.opentelemetry.io/collector/cmd/builder@v${OTEL_VERSION}

RUN GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} builder --config=builder-config.yaml

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build --chmod=755 /app/kyma-otelcol /

USER 65532:65532

ENTRYPOINT ["/kyma-otelcol"]
