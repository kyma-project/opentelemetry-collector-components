package dummymetricsreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"time"
	"go.opentelemetry.io/collector/receiver"
)

type dummyMetricsReceiver struct {
	config       *Config
	nextConsumer consumer.Metrics
	settings     *receiver.CreateSettings

	cancel context.CancelFunc
}

func (r *dummyMetricsReceiver) Start(ctx context.Context, _ component.Host) error {
	ctx = context.Background()
	ctx, r.cancel = context.WithCancel(ctx)

	interval, _ := time.ParseDuration(r.config.Interval)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				r.settings.Logger.Info("DummyMetricsReceiver is generating data")
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (r *dummyMetricsReceiver) Shutdown(ctx context.Context) error {
	if r.cancel != nil {
		r.cancel()
	}
	return nil
}
