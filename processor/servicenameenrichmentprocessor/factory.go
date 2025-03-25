package servicenameenrichmentprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"

	"github.com/kyma-project/opentelemetry-collector-components/processor/serviceenrichmentprocessor/internal/metadata"
)

var processorCapabilities = consumer.Capabilities{MutatesData: true}

func NewFactory() processor.Factory {
	return processor.NewFactory(
		metadata.Type,
		createDefaultConfig,
		processor.WithLogs(createLogsServiceEnrichment, metadata.LogsStability),
		processor.WithTraces(createTracesProcessor, metadata.TracesStability),
		processor.WithMetrics(createMetricsProcessor, metadata.MetricsStability),
	)
}

type Config struct {
	CustomLabels []string `mapstructure:"custom_labels"`
}

func createDefaultConfig() component.Config { return Config{} }

func createLogsServiceEnrichment(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Logs,
) (processor.Logs, error) {
	proc := newServiceEnrichmentProcessor(set.Logger, cfg.(Config))
	return processorhelper.NewLogs(
		ctx,
		set,
		cfg,
		nextConsumer,
		proc.processLogs,
		processorhelper.WithCapabilities(processorCapabilities))
}

func createTracesProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Traces,
) (processor.Traces, error) {
	proc := newServiceEnrichmentProcessor(set.Logger, cfg.(Config))
	return processorhelper.NewTraces(
		ctx,
		set,
		cfg,
		nextConsumer,
		proc.processTraces,
		processorhelper.WithCapabilities(processorCapabilities))
}

func createMetricsProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (processor.Metrics, error) {
	proc := newServiceEnrichmentProcessor(set.Logger, cfg.(Config))
	return processorhelper.NewMetrics(
		ctx,
		set,
		cfg,
		nextConsumer,
		proc.processMetrics,
		processorhelper.WithCapabilities(processorCapabilities))
}
