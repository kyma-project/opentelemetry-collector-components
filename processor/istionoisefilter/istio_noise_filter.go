package istionoisefilter

import (
	"context"
	"regexp"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type istioNoiseFilter struct {
	cfg *Config
}

func newProcessor(cfg *Config) *istioNoiseFilter {
	return &istioNoiseFilter{
		cfg: cfg,
	}
}

type spanAttrs struct {
	namespace        string
	component        string
	canonicalService string
	httpMethod       string
	httpURL          string
	upstreamCluster  string
	userAgent        string
}

func extractSpanAttrs(span ptrace.Span, resourceAttrs pcommon.Map) spanAttrs {
	ns := getStringAttrOrEmpty(resourceAttrs, "k8s.namespace.name")
	spanAttrsMap := span.Attributes()
	return spanAttrs{
		namespace:        ns,
		component:        getStringAttrOrEmpty(spanAttrsMap, "component"),
		canonicalService: getStringAttrOrEmpty(spanAttrsMap, "istio.canonical_service"),
		httpMethod:       getStringAttrOrEmpty(spanAttrsMap, "http.method"),
		httpURL:          getStringAttrOrEmpty(spanAttrsMap, "http.url"),
		upstreamCluster:  getStringAttrOrEmpty(spanAttrsMap, "upstream_cluster.name"),
		userAgent:        getStringAttrOrEmpty(spanAttrsMap, "user_agent"),
	}
}

func (f *istioNoiseFilter) processTraces(_ context.Context, traces ptrace.Traces) (ptrace.Traces, error) {
	for i := range traces.ResourceSpans().Len() {
		resourceSpans := traces.ResourceSpans().At(i)

		for j := range resourceSpans.ScopeSpans().Len() {
			scopeSpans := resourceSpans.ScopeSpans().At(j)

			spans := scopeSpans.Spans()
			spans.RemoveIf(func(span ptrace.Span) bool {
				return shouldFilterSpan(span, resourceSpans.Resource().Attributes())
			})
		}
	}

	return traces, nil
}

var (
	regexHealthzURL          = regexp.MustCompile(`^https://healthz\..+/healthz/ready`)
	regexTelemetryGatewayURL = regexp.MustCompile(`^https?://telemetry-otlp-(logs|metrics|traces)\.kyma-system(\..*)?:(4317|4318).*`)
)

func shouldFilterSpan(span ptrace.Span, resourceAttrs pcommon.Map) bool {
	attrs := extractSpanAttrs(span, resourceAttrs)

	isIstioProxy := attrs.component == "proxy"
	if !isIstioProxy {
		return false
	}

	switch {
	case isTelemetryModuleComponentSpan(attrs):
		return true
	case isAvailabilityServiceProbeSpan(attrs):
		return true
	case isTelemetryGatewaySpan(attrs):
		return true
	case isVictoriaMetricsScrapeSpan(attrs):
		return true
	case isMetricAgentScrapeSpan(attrs):
		return true
	default:
		return false
	}
}

func getStringAttrOrEmpty(attrs pcommon.Map, key string) string {
	attr, ok := attrs.Get(key)
	if !ok {
		return ""
	}

	return attr.Str()
}

// check if the span is from a telemetry module component.
func isTelemetryModuleComponentSpan(attrs spanAttrs) bool {
	if attrs.namespace != "kyma-system" {
		return false
	}

	switch attrs.canonicalService {
	case
		"telemetry-fluent-bit",
		"telemetry-log-agent",
		"telemetry-log-gateway",
		"telemetry-metric-gateway",
		"telemetry-metric-agent",
		"telemetry-trace-gateway":
		return true
	default:
		return false
	}
}

// check if the span is from the availability service probe.
// availability service probes health and readiness endpoints of the istio-ingressgateway.
func isAvailabilityServiceProbeSpan(attrs spanAttrs) bool {
	if attrs.namespace != "istio-system" {
		return false
	}

	if attrs.canonicalService != "istio-ingressgateway" {
		return false
	}

	if attrs.httpMethod != "GET" {
		return false
	}

	if !hasOutboundClusterPrefix(attrs.upstreamCluster) {
		return false
	}

	return regexHealthzURL.MatchString(attrs.httpURL)
}

func hasInboundClusterPrefix(cluster string) bool {
	return len(cluster) >= 8 && cluster[:8] == "inbound|"
}

func hasOutboundClusterPrefix(cluster string) bool {
	return len(cluster) >= 9 && cluster[:9] == "outbound|"
}

// check if the span is from user application to the telemetry gateway.
func isTelemetryGatewaySpan(attrs spanAttrs) bool {
	if attrs.httpMethod != "POST" {
		return false
	}

	if !hasOutboundClusterPrefix(attrs.upstreamCluster) {
		return false
	}

	return regexTelemetryGatewayURL.MatchString(attrs.httpURL)
}

// check if the span is emitted by the VictoriaMetrics(RMA) scraper.
func isVictoriaMetricsScrapeSpan(attrs spanAttrs) bool {
	if attrs.httpMethod != "GET" {
		return false
	}

	if !hasInboundClusterPrefix(attrs.upstreamCluster) {
		return false
	}

	return hasVictoriaMetricsPromscrapeUAPrefix(attrs.userAgent)
}

func hasVictoriaMetricsPromscrapeUAPrefix(ua string) bool {
	return len(ua) >= len("vm_promscrape") && ua[:len("vm_promscrape")] == "vm_promscrape"
}

// check if the span is emitted by the metric agent scraping a user application.
func isMetricAgentScrapeSpan(attrs spanAttrs) bool {
	if attrs.httpMethod != "GET" {
		return false
	}

	if !hasInboundClusterPrefix(attrs.upstreamCluster) {
		return false
	}

	return hasKymaOtelcolUAPrefix(attrs.userAgent)
}

func hasKymaOtelcolUAPrefix(ua string) bool {
	return len(ua) >= len("kyma-otelcol/") && ua[:len("kyma-otelcol/")] == "kyma-otelcol/"
}

func (f *istioNoiseFilter) processMetrics(_ context.Context, metrics pmetric.Metrics) (pmetric.Metrics, error) {
	return pmetric.Metrics{}, nil
}

func (f *istioNoiseFilter) processLogs(_ context.Context, logs plog.Logs) (plog.Logs, error) {
	return plog.Logs{}, nil
}
