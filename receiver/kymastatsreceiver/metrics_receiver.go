package kymastatsreceiver

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"

	operatorv1alpha1 "github.com/kyma-project/telemetry-manager/apis/operator/v1alpha1"
)

type kymaStatsReceiver struct {
	config       *Config
	nextConsumer consumer.Metrics
	settings     *receiver.CreateSettings

	cancel context.CancelFunc
}

func (r *kymaStatsReceiver) Start(ctx context.Context, _ component.Host) error {
	ctx, r.cancel = context.WithCancel(ctx)

	interval := r.config.CollectionInterval
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				md, err := r.pullMetrics(ctx)
				if err != nil {
					r.settings.Logger.Error("Failed to generate metric", zap.Error(err))
					continue
				}
				// nolint:errcheck //
				r.nextConsumer.ConsumeMetrics(ctx, md)
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (r *kymaStatsReceiver) Shutdown(ctx context.Context) error {
	if r.cancel != nil {
		r.cancel()
	}
	return nil
}

func (r *kymaStatsReceiver) pullMetrics(ctx context.Context) (pmetric.Metrics, error) {

	result := r.getTelemetryResources(ctx)
	// kyma_telemetry_status_conditions{version="v1alpha1", type="LogComponentsHealthy", reason="Running"} = 1
	md := pmetric.NewMetrics()
	for _, telemetry := range result.Items {

		resourceMetrics := md.ResourceMetrics().AppendEmpty()
		resourceMetrics.Resource().Attributes().PutStr("version", telemetry.APIVersion)
		resourceMetrics.Resource().Attributes().PutStr("state", string(telemetry.Status.State))
		metric := resourceMetrics.
			ScopeMetrics().
			AppendEmpty().
			Metrics().
			AppendEmpty()

		metric.SetName("kyma.telemetry.status.state")
		gauge := metric.SetEmptyGauge()
		dp := gauge.DataPoints().AppendEmpty()
		dp.SetIntValue(int64(1))

		for _, con := range telemetry.Status.Conditions {
			condMetrics := md.ResourceMetrics().AppendEmpty()
			condMetrics.Resource().Attributes().PutStr("version", telemetry.APIVersion)
			condMetrics.Resource().Attributes().PutStr("type", con.Type)
			condMetrics.Resource().Attributes().PutStr("reason", con.Reason)
			condMetric := condMetrics.
				ScopeMetrics().
				AppendEmpty().
				Metrics().
				AppendEmpty()

			condMetric.SetName("kyma.telemetry.status.conditions")
			condGauge := metric.SetEmptyGauge()
			cdp := condGauge.DataPoints().AppendEmpty()
			cdp.SetIntValue(int64(1))
		}

	}
	return md, nil
}

func (r *kymaStatsReceiver) getTelemetryResources(ctx context.Context) operatorv1alpha1.TelemetryList {
	result := operatorv1alpha1.TelemetryList{}
	c, err := r.config.getK8sClient()
	if err != nil {
		r.settings.Logger.Error("Failed to get k8s client resource", zap.Error(err))
		return result
	}

	err = c.Discovery().RESTClient().Get().Resource("telemetries").Namespace("kyma-system").Do(ctx).Into(&result)
	if err != nil {
		r.settings.Logger.Error("Failed to get telemetry resource", zap.Error(err))
		return result
	}

	return result
}
