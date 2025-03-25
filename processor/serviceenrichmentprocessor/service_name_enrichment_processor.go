package serviceenrichmentprocessor

import (
	"context"
	"regexp"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

var unknownServiceRegex = regexp.MustCompile("^unknown_service(:.+)?$")
var defaultPriority = []string{
	"k8s.deployment.name",
	"k8s.daemonset.name",
	"k8s.statefulset.name",
	"k8s.job.name",
	"k8s.pod.name",
}

type serviceNameEnrichmentProcessor struct {
	logger *zap.Logger
	keys   []string
}

func newServiceNameEnrichmentProcessor(logger *zap.Logger, cfg Config) *serviceNameEnrichmentProcessor {
	keys := cfg.CustomLabels
	keys = append(append(keys, cfg.CustomLabels...), defaultPriority...)
	return &serviceNameEnrichmentProcessor{
		logger: logger,
		keys:   keys,
	}
}

func (sep *serviceNameEnrichmentProcessor) processTraces(ctx context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	res := td.ResourceSpans()
	for i := 0; i < res.Len(); i++ {
		attr := res.At(i).Resource().Attributes()
		sep.setServiceName(attr)
	}
	return td, nil
}

func (sep *serviceNameEnrichmentProcessor) processMetrics(ctx context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	res := md.ResourceMetrics()
	for i := 0; i < res.Len(); i++ {
		attr := res.At(i).Resource().Attributes()
		sep.setServiceName(attr)
	}
	return md, nil
}

func (sep *serviceNameEnrichmentProcessor) processLogs(ctx context.Context, ld plog.Logs) (plog.Logs, error) {
	res := ld.ResourceLogs()
	for i := 0; i < res.Len(); i++ {
		attr := res.At(i).Resource().Attributes()
		sep.setServiceName(attr)

	}
	return ld, nil
}

func (sep *serviceNameEnrichmentProcessor) setServiceName(attr pcommon.Map) {
	svcName, ok := attr.Get("service.name")

	// If service name is set and not unknown return early
	if ok && svcName.AsString() != "" && !unknownServiceRegex.MatchString(svcName.AsString()) {
		return
	}

	// fetch the first svcName available
	svcNameToSet := sep.fetchFirstAvailableServiceName(attr)
	attr.PutStr("service.name", svcNameToSet)
}

func (sep *serviceNameEnrichmentProcessor) fetchFirstAvailableServiceName(attr pcommon.Map) string {
	for _, key := range sep.keys {
		if svcName, ok := attr.Get(key); ok {
			return svcName.AsString()
		}
	}
	return "unknown_service"
}
