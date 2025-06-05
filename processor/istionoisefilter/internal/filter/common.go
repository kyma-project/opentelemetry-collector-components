package filter

import "strings"

func isMetricAgentUserAgent(userAgent string) bool {
	return strings.HasPrefix(userAgent, "kyma-otelcol/")
}

func isRMAUserAgent(userAgent string) bool {
	return strings.HasPrefix(userAgent, "vm_promscrape")
}
