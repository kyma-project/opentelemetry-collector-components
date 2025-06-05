package filter

import (
	"regexp"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

var (
	regexTelemetryGatewayURL = regexp.MustCompile(`^telemetry-otlp-(traces|metrics|logs)\.kyma-system.*`)
	regexDeploymentGateway   = regexp.MustCompile(`^telemetry-(metric|log|trace)-gateway$`)
	regexDaemonsetAgent      = regexp.MustCompile(`^telemetry-(metric-agent|log-agent|fluent-bit)$`)
	regexHealthzDomain       = regexp.MustCompile(`^healthz\..+`)
	regexHealthzPath         = regexp.MustCompile(`/healthz/ready`)
)

type logAttrs struct {
	kymaModule     string
	serverAddress  string
	httpMethod     string
	httpDirection  string
	userAgent      string
	urlPath        string
	resourceNS     string
	resourceDep    string
	resourceDaemon string
}

func extractLogAttrs(log plog.LogRecord, resourceAttrs pcommon.Map) logAttrs {
	attrs := log.Attributes()
	return logAttrs{
		kymaModule:     getStringAttrOrEmpty(attrs, "kyma.module"),
		serverAddress:  getStringAttrOrEmpty(attrs, "server.address"),
		httpMethod:     getStringAttrOrEmpty(attrs, "http.request.method"),
		httpDirection:  getStringAttrOrEmpty(attrs, "http.direction"),
		userAgent:      getStringAttrOrEmpty(attrs, "user_agent.original"),
		urlPath:        getStringAttrOrEmpty(attrs, "url.path"),
		resourceNS:     getStringAttrOrEmpty(resourceAttrs, "k8s.namespace.name"),
		resourceDep:    getStringAttrOrEmpty(resourceAttrs, "k8s.deployment.name"),
		resourceDaemon: getStringAttrOrEmpty(resourceAttrs, "k8s.daemonset.name"),
	}
}

func ShouldDropLogRecord(log plog.LogRecord, resourceAttrs pcommon.Map) bool {
	attrs := extractLogAttrs(log, resourceAttrs)

	if attrs.kymaModule != "istio" {
		return false
	}

	switch {
	case regexTelemetryGatewayURL.MatchString(attrs.serverAddress):
		return true
	case attrs.resourceNS == "kyma-system" && regexDeploymentGateway.MatchString(attrs.resourceDep):
		return true
	case attrs.resourceNS == "kyma-system" && regexDaemonsetAgent.MatchString(attrs.resourceDaemon):
		return true
	case attrs.httpMethod == "GET" && attrs.httpDirection == "inbound" && isRMAUserAgent(attrs.userAgent):
		return true
	case attrs.httpMethod == "GET" && attrs.httpDirection == "inbound" && isMetricAgentUserAgent(attrs.userAgent):
		return true
	case attrs.httpMethod == "GET" && attrs.httpDirection == "outbound" &&
		regexHealthzDomain.MatchString(attrs.serverAddress) &&
		regexHealthzPath.MatchString(attrs.urlPath):
		return true
	default:
		return false
	}
}
