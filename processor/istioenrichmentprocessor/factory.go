package istioenrichmentprocessor

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"

	"github.com/kyma-project/opentelemetry-collector-components/processor/istioenrichmentprocessor/internal/metadata"
)

var processorCapabilities = consumer.Capabilities{MutatesData: true}
var errInvalidConfig = errors.New("invalid configuration")

type Config struct {
	ScopeVersion string `mapstructure:"scope_version"`
}

func createDefaultConfig() component.Config { return Config{} }

func NewFactory() processor.Factory {
	return processor.NewFactory(
		metadata.Type,
		createDefaultConfig,
		processor.WithLogs(createLogsIstioEnrichment, metadata.LogsStability),
	)
}

func createLogsIstioEnrichment(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Logs,
) (processor.Logs, error) {
	config, ok := cfg.(Config)
	if !ok {
		return nil, errInvalidConfig
	}

	proc := newIstioEnrichmentProcessor(set.Logger, config)

	return processorhelper.NewLogs(
		ctx,
		set,
		cfg,
		nextConsumer,
		proc.processLogs,
		processorhelper.WithCapabilities(processorCapabilities))
}
