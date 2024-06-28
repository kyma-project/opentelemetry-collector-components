package components

import (
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/debugexporter"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/receiver"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/dummyreceiver"
)

func Components() (otelcol.Factories, error) {
	var err error
	factories := otelcol.Factories{}

	receivers := []receiver.Factory{
		dummyreceiver.NewFactory(),
	}
	factories.Receivers, err = receiver.MakeFactoryMap(receivers...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	exporters := []exporter.Factory{
		debugexporter.NewFactory(),
	}
	factories.Exporters, err = exporter.MakeFactoryMap(exporters...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	processors := []processor.Factory{
		batchprocessor.NewFactory(),
	}
	factories.Processors, err = processor.MakeFactoryMap(processors...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	return factories, nil
}
