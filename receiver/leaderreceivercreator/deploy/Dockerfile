FROM golang:1.22-alpine as build

RUN apk --update add ca-certificates git

WORKDIR /app
COPY ./ .
RUN ls
RUN go install go.opentelemetry.io/collector/cmd/builder@latest
RUN builder --config=./deploy/builder-config.yaml --skip-strict-versioning

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build --chmod=755 /app/otelcol-dev /

USER 65532:65532

ENTRYPOINT ["/otelcol-dev", "--feature-gates=+component.UseLocalHostAsDefaultHost"]
