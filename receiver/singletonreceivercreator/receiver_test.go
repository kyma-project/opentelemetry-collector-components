package singletonreceivercreator

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"

	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/singletonreceivercreator/internal/metadata"
)

//
//func TestSingletonReceiverCreator(t *testing.T) {
//	var expectedLeaseDurationSeconds = ptr.To(int32(10))
//	config := &Config{
//		leaderElectionConfig: leaderElectionConfig{
//			leaseName:      "my-foo-lease-1",
//			leaseNamespace: "default",
//			leaseDuration:  10 * time.Second,
//			renewDuration:  5 * time.Second,
//			retryPeriod:    2 * time.Second,
//		},
//		subreceiverConfig: receiverConfig{
//			id: component.NewIDWithName(component.MustNewType("dummy"), "name"),
//			config: map[string]any{
//				"interval": "1s",
//			},
//		},
//	}
//	sink := new(consumertest.MetricsSink)
//	// factory := NewFactory()
//	// r, err := factory.CreateMetricsReceiver(context.Background(), receivertest.NewNopSettings(), config, sink)
//	//r := newSingletonReceiverCreator(receivertest.NewNopSettings(), config, sink, "host1")
//
//	telemetryBuilder, err := metadata.NewTelemetryBuilder(componenttest.NewNopTelemetrySettings())
//	require.NoError(t, err)
//
//	r := newSingletonReceiverCreator(
//		receivertest.NewNopSettings(),
//		config,
//		sink,
//		telemetryBuilder,
//		"host1",
//	)
//
//	fakeClient := fake.NewSimpleClientset()
//	config.makeClient = func() (kubernetes.Interface, error) {
//		return fakeClient, nil
//	}
//
//	mh, err := NewMockHost()
//	ctx := context.TODO()
//
//	err = r.Start(ctx, mh)
//	require.NoError(t, err)
//
//	var shutdownOnce sync.Once
//	shutdown := func() {
//		shutdownOnce.Do(func() {
//			assert.NoError(t, r.Shutdown(ctx))
//		})
//	}
//
//	defer shutdown()
//
//	require.Eventually(t, func() bool {
//		lease, err := fakeClient.CoordinationV1().Leases("default").Get(ctx, "my-foo-lease-1", metav1.GetOptions{})
//		require.NoError(t, err)
//		require.NotNil(t, lease)
//		require.Equal(t, expectedLeaseDurationSeconds, lease.Spec.LeaseDurationSeconds)
//		return true
//	}, 5*time.Second, 100*time.Millisecond)
//
//	require.Eventually(t, func() bool {
//		allMetrics := sink.AllMetrics()
//		return len(allMetrics) > 0
//	}, 5*time.Second, 1*time.Second)
//
//	//require.NoError(t, r.Shutdown(ctx))
//}

func TestUnsupportedAuthType(t *testing.T) {
	config := &Config{
		APIConfig: k8sconfig.APIConfig{
			AuthType: "foo",
		}, leaderElectionConfig: leaderElectionConfig{
			leaseName:      "my-foo-lease-1",
			leaseNamespace: "default",
			leaseDuration:  10 * time.Second,
			renewDuration:  5 * time.Second,
			retryPeriod:    2 * time.Second,
		},
		subreceiverConfig: receiverConfig{},
	}

	telemetryBuilder, err := metadata.NewTelemetryBuilder(componenttest.NewNopTelemetrySettings())
	require.NoError(t, err)

	r := newSingletonReceiverCreator(
		receivertest.NewNopSettings(),
		config,
		&consumertest.MetricsSink{},
		telemetryBuilder,
		"host1",
	)

	err = r.Start(context.TODO(), componenttest.NewNopHost())
	require.Error(t, err)

	require.Contains(t, err.Error(), "failed to create Kubernetes client: invalid authType for kubernetes: foo")
}

func TestSingletonReceiverCreator(t *testing.T) {
	var expectedLeaseDurationSeconds = ptr.To(int32(10))

	config := &Config{
		leaderElectionConfig: leaderElectionConfig{
			leaseName:      "my-foo-lease-1",
			leaseNamespace: "default",
			leaseDuration:  10 * time.Second,
			renewDuration:  5 * time.Second,
			retryPeriod:    2 * time.Second,
		},
		subreceiverConfig: receiverConfig{
			id: component.NewIDWithName(component.MustNewType("dummy"), "name"),
			config: map[string]any{
				"interval": "1s",
			},
		},
	}

	ctx := context.TODO()
	fakeClient := fake.NewSimpleClientset()
	config.makeClient = func() (kubernetes.Interface, error) {
		return fakeClient, nil
	}

	telemetryBuilder, err := metadata.NewTelemetryBuilder(componenttest.NewNopTelemetrySettings())
	require.NoError(t, err)

	sink1 := new(consumertest.MetricsSink)
	sr1 := newSingletonReceiverCreator(receivertest.NewNopSettings(), config, sink1, telemetryBuilder, "host1")
	mh1, err := NewMockHost()
	require.NoError(t, err)

	require.NoError(t, sr1.Start(context.TODO(), mh1))

	require.Eventually(t, func() bool {
		lease, err := fakeClient.CoordinationV1().Leases("default").Get(ctx, "my-foo-lease-1", metav1.GetOptions{})
		require.NoError(t, err)
		require.NotNil(t, lease)
		require.Equal(t, expectedLeaseDurationSeconds, lease.Spec.LeaseDurationSeconds)
		return true
	}, 5*time.Second, 100*time.Millisecond)

	require.Eventually(t, func() bool {
		allMetrics := sink1.AllMetrics()
		return len(allMetrics) > 0
	}, 5*time.Second, 1*time.Second)

	sink2 := new(consumertest.MetricsSink)
	sr2 := newSingletonReceiverCreator(receivertest.NewNopSettings(), config, sink2, telemetryBuilder, "host2")
	mh2, err := NewMockHost()
	require.NoError(t, err)

	require.NoError(t, sr2.Start(context.TODO(), mh2))

	require.Never(t, func() bool {
		allMetrics := sink2.AllMetrics()
		return len(allMetrics) > 0
	}, 20*time.Second, 1*time.Second)

	require.NoError(t, sr2.Shutdown(ctx))
	require.NoError(t, sr1.Shutdown(ctx))

}
