package rules

import (
	"regexp"
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

var (
	telemetryModuleGateways = map[string]struct{}{
		"telemetry-log-gateway":    {},
		"telemetry-metric-gateway": {},
		"telemetry-trace-gateway":  {},
	}

	telemetryModuleAgents = map[string]struct{}{
		"telemetry-log-agent":    {},
		"telemetry-metric-agent": {},
		"telemetry-fluent-bit":   {},
	}

	telemetryModuleComponents = mergeSets(
		telemetryModuleGateways,
		telemetryModuleAgents,
	)

	regexTelemetryGatewayURL  = regexp.MustCompile(`^https?://telemetry-otlp-(logs|metrics|traces)\.kyma-system(\..*)?:(4317|4318).*`)
	regexTelemetryGatewayHost = regexp.MustCompile(`^telemetry-otlp-(logs|metrics|traces)\.kyma-system.*`)

	regexHealthzURL  = regexp.MustCompile(`^https://healthz\..+/healthz/ready`)
	regexHealthzHost = regexp.MustCompile(`^healthz\..+`)
	regexHealthzPath = regexp.MustCompile(`/healthz/ready`)
)

func getStringAttrOrEmpty(attrs pcommon.Map, key string) string {
	attr, ok := attrs.Get(key)
	if !ok {
		return ""
	}

	return attr.Str()
}

func mergeSets(a, b map[string]struct{}) map[string]struct{} {
	merged := make(map[string]struct{})
	for k := range a {
		merged[k] = struct{}{}
	}
	for k := range b {
		merged[k] = struct{}{}
	}
	return merged
}

// metric agent proxy scrape spans and access logs can be identified by the user agent
// the user agent is by default set to the name of the collector binary, which is "kyma-otelcol"
func isMetricAgentUserAgent(userAgent string) bool {
	return strings.HasPrefix(userAgent, "kyma-otelcol/")
}

// rma scrape spans and access logs can be identified by the user agent
// the user agent is by default set to "vm_promscrape" (since RMA is based on vmagent)
func isRMAUserAgent(userAgent string) bool {
	return strings.HasPrefix(userAgent, "vm_promscrape")
}
