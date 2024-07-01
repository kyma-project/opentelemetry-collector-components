package singletonreceivercreator

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/receiver/receivertest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"

	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"
)

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
			id: component.NewIDWithName(component.MustNewType("dummymetricsreceiver"), "name"),
			config: map[string]any{
				"interval": "1m",
			},
		},
	}
	r := newSingletonReceiverCreator(receivertest.NewNopSettings(), config, nil, "host1")
	fakeClient := fake.NewSimpleClientset()
	config.makeClient = func() (kubernetes.Interface, error) {
		return fakeClient, nil
	}

	ctx := context.TODO()
	err := r.Start(ctx, componenttest.NewNopHost())
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		lease, err := fakeClient.CoordinationV1().Leases("default").Get(ctx, "my-foo-lease-1", metav1.GetOptions{})
		require.NoError(t, err)
		require.NotNil(t, lease)
		require.Equal(t, expectedLeaseDurationSeconds, lease.Spec.LeaseDurationSeconds)
		return true
	}, 5*time.Second, 100*time.Millisecond)

	require.NoError(t, r.Shutdown(ctx))
}

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
	r := newSingletonReceiverCreator(receivertest.NewNopSettings(), config, nil, "host1")
	err := r.Start(context.TODO(), componenttest.NewNopHost())
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create Kubernetes client: invalid authType for kubernetes: foo")
}

func TestSubReceiverCreation(t *testing.T) {
	var expectedLeaseDurationSeconds = ptr.To(int32(10))

	config := &Config{
		leaderElectionConfig: leaderElectionConfig{
			leaseName:            "my-foo-lease-1",
			leaseNamespace:       "default",
			leaseDurationSeconds: 10 * time.Second,
			renewDeadlineSeconds: 5 * time.Second,
			retryPeriodSeconds:   2 * time.Second,
		},
		subreceiverConfig: receiverConfig{
			id: component.NewIDWithName(component.MustNewType("dummy"), "name"),
			config: map[string]any{
				"interval": "1s",
			},
		},
	}
	sink1 := new(consumertest.MetricsSink)
	fakeClient := fake.NewSimpleClientset()
	config.makeClient = func() (k8s.Interface, error) {
		return fakeClient, nil
	}

	ctx := context.TODO()
	mr1 := newSingletonReceiverCreator(receivertest.NewNopSettings(), config, sink1, "host1")
	mh1, err := NewMockHost()
	require.NoError(t, err)

	require.NoError(t, mr1.Start(context.TODO(), mh1))

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
	mr2 := newSingletonReceiverCreator(receivertest.NewNopSettings(), config, sink2, "host2")
	mh2, err := NewMockHost()
	err = mr2.Start(context.TODO(), mh2)

	assert.Neverf(t, func() bool {
		allMetrics := sink2.AllMetrics()
		return len(allMetrics) != 0
	}, 30*time.Second, 1*time.Second, "metrics received by next consumer")
	require.NoError(t, mr2.Shutdown(ctx))
	require.NoError(t, mr1.Shutdown(ctx))

}
