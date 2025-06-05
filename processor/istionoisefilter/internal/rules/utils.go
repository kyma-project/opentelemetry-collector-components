package rules

import "go.opentelemetry.io/collector/pdata/pcommon"

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
