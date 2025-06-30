package istionoisefilter

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"

	"github.com/kyma-project/opentelemetry-collector-components/processor/istionoisefilter/internal/metadata"
)

var (
	processorCapabilities = consumer.Capabilities{MutatesData: true}
	errInvalidConfig      = errors.New("invalid configuration, expected *istionoisefilter.Config")
)

func createDefaultConfig() component.Config {
	return &Config{}
}

func NewFactory() processor.Factory {
	return processor.NewFactory(
		metadata.Type,
		createDefaultConfig,
		processor.WithLogs(createLogProcessor, metadata.LogsStability),
		processor.WithMetrics(createMetricsProcessor, metadata.MetricsStability),
		processor.WithTraces(createTracesProcessor, metadata.TracesStability),
	)
}

func createLogProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Logs,
) (processor.Logs, error) {
	c, ok := cfg.(*Config)
	if !ok {
		return nil, errInvalidConfig
	}

	proc := newProcessor(c, set)

	return processorhelper.NewLogs(
		ctx,
		set,
		cfg,
		nextConsumer,
		proc.processLogs,
		processorhelper.WithCapabilities(processorCapabilities))
}

func createMetricsProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (processor.Metrics, error) {
	c, ok := cfg.(*Config)
	if !ok {
		return nil, errInvalidConfig
	}

	proc := newProcessor(c, set)

	return processorhelper.NewMetrics(
		ctx,
		set,
		cfg,
		nextConsumer,
		proc.processMetrics,
		processorhelper.WithCapabilities(processorCapabilities))
}

func createTracesProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Traces,
) (processor.Traces, error) {
	c, ok := cfg.(*Config)
	if !ok {
		return nil, errInvalidConfig
	}

	proc := newProcessor(c, set)

	return processorhelper.NewTraces(
		ctx,
		set,
		cfg,
		nextConsumer,
		proc.processTraces,
		processorhelper.WithCapabilities(processorCapabilities))
}
