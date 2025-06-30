package rules

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func ShouldDropSpan(span ptrace.Span, resourceAttrs pcommon.Map) bool {
	attrs := extractSpanAttrs(span, resourceAttrs)

	// component must be "proxy" to be considered an Istio proxy span.
	isIstioProxy := attrs.component == "proxy"
	if !isIstioProxy {
		return false
	}

	switch {
	case isTelemetryModuleComponentSpan(attrs):
		return true
	case isTelemetryGatewaySpan(attrs):
		return true
	case isMetricScrapeSpan(attrs):
		return true
	case isAvailabilityServiceProbeSpan(attrs):
		return true
	default:
		return false
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

// check if the span is from a telemetry module component.
func isTelemetryModuleComponentSpan(attrs spanAttrs) bool {
	if attrs.namespace != "kyma-system" {
		return false
	}

	if _, found := telemetryModuleComponents[attrs.canonicalService]; found {
		return true
	}

	return false
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

func isTelemetryGatewaySpan(attrs spanAttrs) bool {
	if attrs.httpMethod != "POST" {
		return false
	}

	if !hasOutboundClusterPrefix(attrs.upstreamCluster) {
		return false
	}

	return regexTelemetryGatewayURL.MatchString(attrs.httpURL)
}

func isMetricScrapeSpan(attrs spanAttrs) bool {
	if attrs.httpMethod != "GET" {
		return false
	}

	if !hasInboundClusterPrefix(attrs.upstreamCluster) {
		return false
	}

	return isMetricAgentUserAgent(attrs.userAgent) || isRMAUserAgent(attrs.userAgent)
}
