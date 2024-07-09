package dummyreceiver

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

type dummyReceiver struct {
	config       *Config
	nextConsumer consumer.Metrics
	settings     *receiver.Settings

	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func (r *dummyReceiver) Start(_ context.Context, _ component.Host) error { //nolint:contextcheck // Create a new context as specified in the interface documentation
	r.settings.Logger.Info("Starting dummy receiver", zap.String("interval", r.config.Interval))
	ctx := context.Background()
	ctx, r.cancel = context.WithCancel(ctx)

	interval, err := time.ParseDuration(r.config.Interval)
	if err != nil {
		return fmt.Errorf("failed to parse interval: %w", err)

	}

	r.wg.Add(1)
	go r.startGenerating(ctx, interval) //nolint:contextcheck // Non-inherited new context

	return nil
}

func (r *dummyReceiver) startGenerating(ctx context.Context, interval time.Duration) {
	defer r.wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			md, err := r.generateMetric()
			if err != nil {
				r.settings.Logger.Error("Failed to generate metric", zap.Error(err))
				continue
			}
			err = r.nextConsumer.ConsumeMetrics(ctx, md)
			if err != nil {
				r.settings.Logger.Error("next consumer failed", zap.Error(err))
			}
		}
	}
}

func (r *dummyReceiver) generateMetric() (pmetric.Metrics, error) {
	r.settings.Logger.Debug("Generating metric")
	host, err := os.Hostname()
	if err != nil {
		return pmetric.Metrics{}, fmt.Errorf("failed to get hostname: %w", err)
	}

	md := pmetric.NewMetrics()
	resourceMetrics := md.ResourceMetrics().AppendEmpty()
	resourceMetrics.Resource().Attributes().PutStr("k8s.cluster.name", "test-cluster")
	metric := resourceMetrics.
		ScopeMetrics().
		AppendEmpty().
		Metrics().
		AppendEmpty()

	metric.SetName("dummy")
	metric.SetDescription("a dummy gauge")
	gauge := metric.SetEmptyGauge()
	for i := 0; i < 5; i++ {
		dp := gauge.DataPoints().AppendEmpty()
		dp.SetIntValue(int64(i))
		dp.Attributes().PutStr("host", host)
	}

	return md, nil
}

func (r *dummyReceiver) Shutdown(_ context.Context) error {
	r.settings.Logger.Info("Shutting down dummy receiver")
	if r.cancel != nil {
		r.cancel()
	}
	r.wg.Wait()
	return nil
}
