package dummyreceiver

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/kyma-project/opentelemetry-collector-components/extension/leaderelector"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

type leaderElectionReceiver struct {
	Id component.ID `mapstructure:"name"`
}

type dummyReceiver struct {
	config       *Config
	nextConsumer consumer.Metrics
	settings     *receiver.Settings
	startCh      chan struct{}

	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func (ler *leaderElectionReceiver) getExtensionClient(extensions map[component.ID]component.Component) (component.Component, error) {
	if ext, found := extensions[ler.Id]; found {
		return ext, nil
	}
	return nil, errors.New("extension not found")

}

func (r *dummyReceiver) startReceiver(ctx context.Context) error {
	// Do something
	r.settings.Logger.Info("Starting dummy receiver", zap.String("interval", r.config.Interval))
	interval, err := time.ParseDuration(r.config.Interval)
	if err != nil {
		return fmt.Errorf("failed to parse interval: %w", err)

	}
	r.wg.Add(1)
	go r.startGenerating(ctx, interval) //nolint:contextcheck // Non-inherited new context
	return nil
}

func (r *dummyReceiver) stopReceiver() error {
	r.settings.Logger.Info("Shutting down dummy receiver")
	if r.cancel != nil {
		r.cancel()
	}
	r.wg.Wait()
	return nil
}

func (r *dummyReceiver) fetchAndCheckExtension(host component.Host) (leaderelector.LeaderElection, error) {
	extList := host.GetExtensions()
	if extList == nil {
		return nil, errors.New("no extensions found")
	}
	r.settings.Logger.Info("Leader election enabled")
	ext := extList[component.ID(r.config.LeaseName.Id)]
	if ext == nil {
		return nil, errors.New("extension not found")
	}
	leaderElectionExtension := ext.(leaderelector.LeaderElection)
	return leaderElectionExtension, nil
}

func (r *dummyReceiver) Start(_ context.Context, host component.Host) error { //nolint:contextcheck // Create a new context as specified in the interface documentation
	ctx := context.Background()
	ctx, r.cancel = context.WithCancel(ctx)
	if r.config.LeaseName != nil {
		leaderelectorExt, err := r.fetchAndCheckExtension(host)
		if err != nil {
			return err
		}
		leaderelectorExt.SetCallBackFuncs(
			func(ctx context.Context) {
				if err := r.startReceiver(context.TODO()); err != nil {
					r.settings.Logger.Error("Failed to start receiver", zap.Error(err))
				}
			}, func() {
				if err := r.stopReceiver(); err != nil {
					r.settings.Logger.Error("Failed to stop receiver", zap.Error(err))
				}
			},
		)
	} else {
		if err := r.startReceiver(ctx); err != nil {
			return err
		}
	}

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
	if r.config.LeaseName == nil {
		if err := r.stopReceiver(); err != nil {
			return err
		}
	}
	return nil
}
