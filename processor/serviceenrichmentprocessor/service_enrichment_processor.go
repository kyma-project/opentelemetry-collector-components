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

type serviceEnrichmentProcessor struct {
	logger *zap.Logger
}

func (sep *serviceEnrichmentProcessor) processTraces(ctx context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	res := td.ResourceSpans()
	for i := 0; i < res.Len(); i++ {
		attr := res.At(i).Resource().Attributes()
		sep.setServiceName(attr)
	}
	return td, nil
}

func (sep *serviceEnrichmentProcessor) processMetrics(ctx context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	res := md.ResourceMetrics()
	for i := 0; i < res.Len(); i++ {
		attr := res.At(i).Resource().Attributes()
		sep.setServiceName(attr)
	}
	return md, nil
}

func (sep *serviceEnrichmentProcessor) processLogs(ctx context.Context, ld plog.Logs) (plog.Logs, error) {
	res := ld.ResourceLogs()
	for i := 0; i < res.Len(); i++ {
		attr := res.At(i).Resource().Attributes()
		sep.setServiceName(attr)

	}
	return ld, nil
}

func (sep *serviceEnrichmentProcessor) setServiceName(attr pcommon.Map) {
	unknownServiceRegex := regexp.MustCompile("^unknown_service(:.+)?$")
	svcName, ok := attr.Get("service.name")

	// If service name is set and not unknown return early
	if ok && svcName.AsString() != "" && !unknownServiceRegex.MatchString(svcName.AsString()) {
		return
	}

	// fetch the first svcName available
	svcNameToSet := fetchFirstAvailableServiceName(attr)
	attr.PutStr("service.name", svcNameToSet)
}

func fetchFirstAvailableServiceName(attr pcommon.Map) string {
	keys := []string{
		"kyma.kubernetes_io_app_name",
		"kyma.app_name",
		"k8s.deployment.name",
		"k8s.daemonset.name",
		"k8s.statefulset.name",
		"k8s.job.name",
		"k8s.pod.name",
	}

	for _, key := range keys {
		if svcName, ok := attr.Get(key); ok {
			return svcName.AsString()
		}
	}
	return "unknown_service"
}
