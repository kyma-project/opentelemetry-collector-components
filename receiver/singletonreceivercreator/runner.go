// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package singletonreceivercreator

// This code has been copied fron: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/receivercreator/runner.go
// Some modifications have been made to the original code to better suit the needs of this project.
import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

// receiverRunner handles starting/stopping of a concrete wrapped receiver instance.
type receiverRunner struct {
	logger      *zap.Logger
	params      receiver.Settings
	idNamespace component.ID
	host        component.Host
	receiver    component.Component
	lock        *sync.Mutex
}

func newReceiverRunner(params receiver.Settings, host component.Host) *receiverRunner {
	return &receiverRunner{
		logger:      params.Logger,
		params:      params,
		idNamespace: params.ID,
		host:        host,
		lock:        &sync.Mutex{},
	}
}

func (r *receiverRunner) start(config receiverConfig, metricsConsumer consumer.Metrics) error {
	factory := r.host.GetFactory(component.KindReceiver, config.id.Type())

	if factory == nil {
		return fmt.Errorf("unable to lookup factory for wrapped receiver %q", config.id.String())
	}

	receiverFactory, ok := factory.(receiver.Factory)
	if !ok {
		return fmt.Errorf("factory %q is not a receiver factory", config.id.Type())
	}

	cfg, err := r.loadReceiverConfig(receiverFactory, config)
	if err != nil {
		return err
	}

	// Sets dynamically created wrapped receiver to something like receiver_creator/1/redis.
	id := component.NewIDWithName(factory.Type(), fmt.Sprintf("%s/%s", config.id.Name(), r.idNamespace))
	r.logger.Debug("Creating wrapped receiver", zap.String("receiver", id.String()))

	wr := &wrappedReceiver{}
	var createError error

	if wr.metrics, err = r.createMetricsRuntimeReceiver(receiverFactory, id, cfg, metricsConsumer); err != nil {
		if errors.Is(err, component.ErrDataTypeIsNotSupported) {
			r.logger.Info("instantiated receiver doesn't support metrics", zap.String("receiver", config.id.String()), zap.Error(err))
			wr.metrics = nil
		} else {
			createError = multierr.Combine(createError, err)
		}
	}

	if createError != nil {
		return fmt.Errorf("failed creating wrapped receiver: %w", createError)
	}

	r.params.Logger.Debug("Starting wrapped receiver with config", zap.String("receiver", config.id.String()), zap.Any("config", cfg))

	if err = wr.Start(context.Background(), r.host); err != nil {
		return fmt.Errorf("failed starting wrapped receiver: %w", err)
	}

	r.receiver = wr

	return nil
}

// shutdown the given receiver.
func (r *receiverRunner) shutdown(ctx context.Context) error {
	if r.receiver != nil {
		return r.receiver.Shutdown(ctx)
	}
	return nil
}

func (r *receiverRunner) loadReceiverConfig(factory receiver.Factory, receiver receiverConfig) (component.Config, error) {
	receiverCfg := factory.CreateDefaultConfig()
	config := confmap.NewFromStringMap(receiver.config)
	if err := config.Unmarshal(receiverCfg); err != nil {
		return nil, fmt.Errorf("failed to load %q subreceiver config: %w", receiver.id.String(), err)
	}
	return receiverCfg, nil
}

// createMetricsRuntimeReceiver creates a receiver that is discovered at runtime.
func (r *receiverRunner) createMetricsRuntimeReceiver(
	factory receiver.Factory,
	id component.ID,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (receiver.Metrics, error) {
	runParams := r.params
	runParams.Logger = runParams.Logger.With(zap.String("name", id.String()))
	runParams.ID = id
	return factory.CreateMetricsReceiver(context.Background(), runParams, cfg, nextConsumer)
}

var _ component.Component = (*wrappedReceiver)(nil)

type wrappedReceiver struct {
	metrics receiver.Metrics
}

func (w *wrappedReceiver) Start(ctx context.Context, host component.Host) error {
	var err error
	for _, r := range []component.Component{w.metrics} {
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
	for _, r := range []component.Component{w.metrics} {
		if r != nil {
			if e := r.Shutdown(ctx); e != nil {
				err = multierr.Combine(err, e)
			}
		}
	}
	return err
}
