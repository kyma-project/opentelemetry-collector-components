package dummymetricsreceiver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.opentelemetry.io/collector/component/componenttest"
)

func TestReceiveMetrics(t *testing.T) {
	sink := new(consumertest.MetricsSink)

	cfg := &Config{
		Interval: (1 * time.Second).String(),
	}
	mr, err := createMetricsReceiver(context.Background(), receivertest.NewNopCreateSettings(), cfg, sink)
	require.NoError(t, err)

	err = mr.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		allMetrics := sink.AllMetrics()
		return len(allMetrics) > 0
	}, 5*time.Second, 1*time.Second)
}
