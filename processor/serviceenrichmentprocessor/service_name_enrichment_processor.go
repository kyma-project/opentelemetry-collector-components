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

const unknownService = "unknown_service"

var unknownServiceRegex = regexp.MustCompile("^unknown_service(:.+)?$")

var defaultAttributeKeysPriority = []string{
	"k8s.deployment.name",
	"k8s.daemonset.name",
	"k8s.statefulset.name",
	"k8s.job.name",
	"k8s.pod.name",
}

type serviceEnrichmentProcessor struct {
	logger   *zap.Logger
	attrKeys []string
}

func newServiceEnrichmentProcessor(logger *zap.Logger, cfg Config) *serviceEnrichmentProcessor {
	attrKeys := cfg.resourceAttributes
	attrKeys = append(attrKeys, cfg.resourceAttributes...)
	attrKeys = append(attrKeys, defaultAttributeKeysPriority...)

	return &serviceEnrichmentProcessor{
		logger:   logger,
		attrKeys: attrKeys,
	}
}

func (sep *serviceEnrichmentProcessor) processTraces(_ context.Context, traces ptrace.Traces) (ptrace.Traces, error) {
	resourceSpans := traces.ResourceSpans()
	for i := 0; i < resourceSpans.Len(); i++ {
		attributes := resourceSpans.At(i).Resource().Attributes()
		sep.enrichServiceName(attributes)
	}
	return traces, nil
}

func (sep *serviceEnrichmentProcessor) processMetrics(_ context.Context, metrics pmetric.Metrics) (pmetric.Metrics, error) {
	resourceMetrics := metrics.ResourceMetrics()
	for i := 0; i < resourceMetrics.Len(); i++ {
		attr := resourceMetrics.At(i).Resource().Attributes()
		sep.enrichServiceName(attr)
	}
	return metrics, nil
}

func (sep *serviceEnrichmentProcessor) processLogs(_ context.Context, logs plog.Logs) (plog.Logs, error) {
	resourceLogs := logs.ResourceLogs()
	for i := 0; i < resourceLogs.Len(); i++ {
		attr := resourceLogs.At(i).Resource().Attributes()
		sep.enrichServiceName(attr)

	}
	return logs, nil
}

func (sep *serviceEnrichmentProcessor) enrichServiceName(attr pcommon.Map) {
	serviceName, exists := attr.Get("service.name")

	// If service name is set and not unknown return early
	if exists && serviceName.AsString() != "" && !unknownServiceRegex.MatchString(serviceName.AsString()) {
		return
	}

	// fetch the first svcName available
	svcNameToSet := sep.fetchFirstAvailableServiceName(attr)
	attr.PutStr("service.name", svcNameToSet)
}

func (sep *serviceEnrichmentProcessor) fetchFirstAvailableServiceName(attributes pcommon.Map) string {
	for _, key := range sep.attrKeys {
		if serviceName, ok := attributes.Get(key); ok {
			return serviceName.AsString()
		}
	}
	return unknownService
}
