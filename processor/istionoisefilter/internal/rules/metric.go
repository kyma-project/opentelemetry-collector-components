package rules

import (
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

const (
	istioMetricPrefix = "istio_"
)

// ShouldDropMetricDataPoint checks if the given metric is an Istio metric that records communication between telemetry module components,
// or between a telemetry module component and a workload, and should be dropped since it does not provide useful information to the user.
func ShouldDropMetricDataPoint(metricName string, dataPointAttrs pcommon.Map) bool {
	if !strings.HasPrefix(metricName, istioMetricPrefix) {
		return false
	}

	if sourceWorkload, found := dataPointAttrs.Get("source_workload"); found && sourceWorkload.Str() == "telemetry-metric-agent" {
		return true
	}

	if destinationWorkload, found := dataPointAttrs.Get("destination_workload"); found {
		// check if the destination workload is one of the telemetry module gateways
		// since only gateways can be one the receiving side
		if _, found := telemetryModuleGateways[destinationWorkload.Str()]; found {
			return true
		}
	}

	return false
}
