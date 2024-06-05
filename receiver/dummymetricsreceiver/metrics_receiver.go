package dummymetricsreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
)

type dummyMetricsReceiver struct {
	host   component.Host
	cancel context.CancelFunc
}

func (r *dummyMetricsReceiver) Start(ctx context.Context, host component.Host) error {
	r.host = host
	ctx = context.Background()
	ctx, r.cancel = context.WithCancel(ctx)

	return nil
}

func (r *dummyMetricsReceiver) Shutdown(ctx context.Context) error {
	if r.cancel != nil {
		r.cancel()
	}
	return nil
}
