package dummymetricsreceiver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"
)

func TestReceiveMetrics(t *testing.T) {
	sink := new(consumertest.MetricsSink)

	cfg := &Config{
		Interval: "1s",
	}
	mr, err := createMetricsReceiver(context.Background(), receivertest.NewNopSettings(), cfg, sink)
	require.NoError(t, err)

	err = mr.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		allMetrics := sink.AllMetrics()
		return len(allMetrics) > 0
	}, 5*time.Second, 1*time.Second)
	err = mr.Shutdown(context.Background())
	require.NoError(t, err)
}
