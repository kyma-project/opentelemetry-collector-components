// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package leaderreceivercreator

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/consumer"
	rcvr "go.opentelemetry.io/collector/receiver"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

// receiverRunner handles starting/stopping of a concrete subreceiver instance.
type receiverRunner struct {
	logger      *zap.Logger
	params      rcvr.CreateSettings
	idNamespace component.ID
	host        component.Host
	receiver    component.Component
	lock        *sync.Mutex
}

func newReceiverRunner(params rcvr.CreateSettings, host component.Host) *receiverRunner {
	return &receiverRunner{
		logger:      params.Logger,
		params:      params,
		idNamespace: params.ID,
		host:        host,
		lock:        &sync.Mutex{},
	}
}

func (run *receiverRunner) start(
	receiver receiverConfig,
	metricsConsumer consumer.Metrics,
) error {
	factory := run.host.GetFactory(component.KindReceiver, receiver.id.Type())

	if factory == nil {
		return fmt.Errorf("unable to lookup factory for receiver %q", receiver.id.String())
	}

	receiverFactory := factory.(rcvr.Factory)

	cfg, _, err := run.loadReceiverConfig(receiverFactory, receiver)
	if err != nil {
		return err
	}

	// Sets dynamically created receiver to something like receiver_creator/1/redis.
	id := component.NewIDWithName(factory.Type(), fmt.Sprintf("%s/%s", receiver.id.Name(), run.idNamespace))

	wr := &wrappedReceiver{}
	var createError error

	if wr.metrics, err = run.createMetricsRuntimeReceiver(receiverFactory, id, cfg, metricsConsumer); err != nil {
		if errors.Is(err, component.ErrDataTypeIsNotSupported) {
			run.logger.Info("instantiated receiver doesn't support metrics", zap.String("receiver", receiver.id.String()), zap.Error(err))
			wr.metrics = nil
		} else {
			createError = multierr.Combine(createError, err)
		}
	}

	if createError != nil {
		return fmt.Errorf("failed creating endpoint-derived receiver: %w", createError)
	}

	run.params.Logger.Info("Starting subreceiver",
		zap.String("receiver", receiver.id.String()),
		zap.Any("config", cfg))

	if err = wr.Start(context.Background(), run.host); err != nil {
		return fmt.Errorf("failed starting endpoint-derived receiver: %w", err)
	}

	run.receiver = wr

	return nil
}

// shutdown the given receiver.
func (run *receiverRunner) shutdown(ctx context.Context) error {
	if run.receiver != nil {
		return run.receiver.Shutdown(ctx)
	}
	return nil
}

func (run *receiverRunner) loadReceiverConfig(
	factory rcvr.Factory,
	receiver receiverConfig,
) (component.Config, string, error) {
	receiverCfg := factory.CreateDefaultConfig()
	if err := component.UnmarshalConfig(confmap.NewFromStringMap(receiver.config), receiverCfg); err != nil {
		return nil, "", fmt.Errorf("failed to load %q subreceiver config: %w", receiver.id.String(), err)
	}
	return receiverCfg, "", nil
}

// createLogsRuntimeReceiver creates a receiver that is discovered at runtime.
func (run *receiverRunner) createLogsRuntimeReceiver(
	factory rcvr.Factory,
	id component.ID,
	cfg component.Config,
	nextConsumer consumer.Logs,
) (rcvr.Logs, error) {
	runParams := run.params
	runParams.Logger = runParams.Logger.With(zap.String("name", id.String()))
	runParams.ID = id
	return factory.CreateLogsReceiver(context.Background(), runParams, cfg, nextConsumer)
}

// createMetricsRuntimeReceiver creates a receiver that is discovered at runtime.
func (run *receiverRunner) createMetricsRuntimeReceiver(
	factory rcvr.Factory,
	id component.ID,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (rcvr.Metrics, error) {
	runParams := run.params
	runParams.Logger = runParams.Logger.With(zap.String("name", id.String()))
	runParams.ID = id
	return factory.CreateMetricsReceiver(context.Background(), runParams, cfg, nextConsumer)
}

// createTracesRuntimeReceiver creates a receiver that is discovered at runtime.
func (run *receiverRunner) createTracesRuntimeReceiver(
	factory rcvr.Factory,
	id component.ID,
	cfg component.Config,
	nextConsumer consumer.Traces,
) (rcvr.Traces, error) {
	runParams := run.params
	runParams.Logger = runParams.Logger.With(zap.String("name", id.String()))
	runParams.ID = id
	return factory.CreateTracesReceiver(context.Background(), runParams, cfg, nextConsumer)
}

var _ component.Component = (*wrappedReceiver)(nil)

type wrappedReceiver struct {
	logs    rcvr.Logs
	metrics rcvr.Metrics
	traces  rcvr.Traces
}

func (w *wrappedReceiver) Start(ctx context.Context, host component.Host) error {
	var err error
	for _, r := range []component.Component{w.logs, w.metrics, w.traces} {
		if r != nil {
			if e := r.Start(ctx, host); e != nil {
				err = multierr.Combine(err, e)
			}
		}
	}
	return err
}

func (w *wrappedReceiver) Shutdown(ctx context.Context) error {
	var err error
	for _, r := range []component.Component{w.logs, w.metrics, w.traces} {
		if r != nil {
			if e := r.Shutdown(ctx); e != nil {
				err = multierr.Combine(err, e)
			}
		}
	}
	return err
}
