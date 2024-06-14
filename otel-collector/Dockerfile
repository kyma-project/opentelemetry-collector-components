ARG GOLANG_VERSION

FROM golang:${GOLANG_VERSION}-alpine as build
ARG OTEL_VERSION
ARG KYMA_OCC_VERSION
ARG OTEL_CONTRIB_VERSION=${OTEL_VERSION}

ARG BUILD_MODE="PR"

RUN apk --update add ca-certificates git

ADD receiver /receiver/
WORKDIR /app
COPY otel-collector/builder-config-${BUILD_MODE}.yaml builder-config.yaml

ENV OTEL_VERSION ${OTEL_VERSION}
ENV OTEL_CONTRIB_VERSION ${OTEL_CONTRIB_VERSION}
RUN sed -i s/OTEL_VERSION/${OTEL_VERSION}/g builder-config.yaml
RUN sed -i s/OTEL_CONTRIB_VERSION/${OTEL_CONTRIB_VERSION}/g builder-config.yaml
RUN sed -i s/KYMA_OCC_VERSION/${KYMA_OCC_VERSION}/g builder-config.yaml

RUN go install go.opentelemetry.io/collector/cmd/builder@v${OTEL_VERSION}

RUN builder --config=builder-config.yaml

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build --chmod=755 /app/kyma-otelcol /

USER 65532:65532

ENTRYPOINT ["/kyma-otelcol", "--feature-gates=+component.UseLocalHostAsDefaultHost"]
