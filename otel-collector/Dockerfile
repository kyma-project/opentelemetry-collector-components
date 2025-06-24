FROM --platform=$BUILDPLATFORM golang:1.24.4-alpine3.21 AS otel-builder-tool

ARG OTEL_VERSION
RUN go install go.opentelemetry.io/collector/cmd/builder@v${OTEL_VERSION}

FROM --platform=$BUILDPLATFORM golang:1.24.4-alpine3.21 AS builder
ARG TARGETOS
ARG TARGETARCH
ARG OTEL_VERSION
ARG OTEL_CONTRIB_VERSION=${OTEL_VERSION}

COPY --from=otel-builder-tool /go/bin/builder /usr/local/bin/builder

WORKDIR /app

COPY go.mod go.sum ./
#COPY otel-collector/builder-config.yaml builder-config.yaml
#ADD receiver ./receiver/
#ADD internal ./internal/
#ADD processor ./processor/

RUN go mod download

COPY otel-collector/builder-config.yaml builder-config.yaml

ENV OTEL_VERSION=${OTEL_VERSION}
ENV OTEL_CONTRIB_VERSION=${OTEL_CONTRIB_VERSION}

RUN apk --update add --no-cache ca-certificates git && \
    sed -i "s/OTEL_VERSION/${OTEL_VERSION}/g" builder-config.yaml && \
    sed -i "s/OTEL_CONTRIB_VERSION/${OTEL_CONTRIB_VERSION}/g" builder-config.yaml && \
    GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} builder --config=builder-config.yaml

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder --chmod=755 /app/kyma-otelcol /

USER 65532:65532

ENTRYPOINT ["/kyma-otelcol"]


#FROM --platform=$BUILDPLATFORM golang:1.24.4-alpine3.21 AS deps
#
#ARG OTEL_VERSION
#
#WORKDIR /app
#
#COPY otel-collector/builder-config.yaml builder-config.yaml
#COPY go.mod go.sum ./
#
#ADD receiver /app/receiver/
#ADD internal /app/internal/
#ADD processor /app/processor/
#
#RUN apk --update add --no-cache git
#RUN go install go.opentelemetry.io/collector/cmd/builder@v${OTEL_VERSION}
#
#ARG OTEL_CONTRIB_VERSION=${OTEL_VERSION}
#RUN sed -i "s/OTEL_VERSION/${OTEL_VERSION}/g" builder-config.yaml && \
#    sed -i "s/OTEL_CONTRIB_VERSION/${OTEL_CONTRIB_VERSION}/g" builder-config.yaml
#RUN builder --config=builder-config.yaml --skip-compilation && go mod download
#
#FROM --platform=$BUILDPLATFORM golang:1.24.4-alpine3.21 AS build
#
#ARG TARGETOS
#ARG TARGETARCH
#ARG OTEL_VERSION
#ARG OTEL_CONTRIB_VERSION=${OTEL_VERSION}
#
#RUN apk --update add ca-certificates git
#WORKDIR /app
#
#COPY --from=deps /go/pkg/mod /go/pkg/mod
#COPY --from=deps /root/.cache/go-build /root/.cache/go-build
#COPY --from=deps /app .
#COPY --from=deps /go/bin/builder /usr/local/bin/builder
#
#RUN GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} builder --config=builder-config.yaml
#
#FROM scratch
#
#COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
#COPY --from=build --chmod=755 /app/kyma-otelcol /
#
#USER 65532:65532
#
#ENTRYPOINT ["/kyma-otelcol"]