package singletonreceivercreator

import (
	"context"
	"testing"
	"time"

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
		subreceiverConfig: receiverConfig{},
	}
	r, err := newSingletonReceiverCreator(receivertest.NewNopSettings(), config)
	require.NoError(t, err)

	fakeClient := fake.NewSimpleClientset()
	config.makeClient = func() (kubernetes.Interface, error) {
		return fakeClient, nil
	}

	ctx := context.TODO()
	err = r.Start(ctx, componenttest.NewNopHost())
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
	r, err := newSingletonReceiverCreator(receivertest.NewNopSettings(), config)
	require.NoError(t, err)

	err = r.Start(context.TODO(), componenttest.NewNopHost())
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create Kubernetes client: invalid authType for kubernetes: foo")
}
