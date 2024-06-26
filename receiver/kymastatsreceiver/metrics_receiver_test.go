package kymastatsreceiver

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
		CollectionInterval: 1 * time.Second,
	}
	mr, err := createMetricsReceiver(context.Background(), receivertest.NewNopCreateSettings(), cfg, sink)
	require.NoError(t, err)

	err = mr.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	err = mr.Shutdown(context.Background())
	require.NoError(t, err)
}
