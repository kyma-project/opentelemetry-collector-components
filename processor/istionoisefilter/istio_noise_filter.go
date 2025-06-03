package istionoisefilter

import (
	"context"

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

func (f *istioNoiseFilter) processTraces(_ context.Context, traces ptrace.Traces) (ptrace.Traces, error) {
	return ptrace.Traces{}, nil
}

func (f *istioNoiseFilter) processMetrics(_ context.Context, metrics pmetric.Metrics) (pmetric.Metrics, error) {
	return pmetric.Metrics{}, nil
}

func (f *istioNoiseFilter) processLogs(_ context.Context, logs plog.Logs) (plog.Logs, error) {
	return plog.Logs{}, nil
}
