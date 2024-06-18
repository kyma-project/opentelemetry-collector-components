package leaderreceivercreator

import (
	"context"
	"k8s.io/utils/ptr"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/receiver/receivertest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestMockReceiverCreator(t *testing.T) {
	var expectedLeaseDurationSeconds = ptr.To(int32(10))
	config := &Config{
		leaderElectionConfig: leaderElectionConfig{
			leaseName:            "my-foo-lease-1",
			leaseNamespace:       "default",
			leaseDurationSeconds: 10 * time.Second,
			renewDeadlineSeconds: 5 * time.Second,
			retryPeriodSeconds:   2 * time.Second,
		},
		subreceiverConfig: receiverConfig{},
	}
	r := newLeaderReceiverCreator(receivertest.NewNopCreateSettings(), config)
	lr := r.(*leaderReceiverCreator)
	fakeClient := fake.NewSimpleClientset()
	lr.makeClient = func() (kubernetes.Interface, error) {
		return fakeClient, nil
	}

	ctx := context.TODO()
	lr.Start(ctx, componenttest.NewNopHost())
	require.Eventually(t, func() bool {
		lease, err := fakeClient.CoordinationV1().Leases("default").Get(ctx, "my-foo-lease-1", metav1.GetOptions{})
		require.NoError(t, err)
		require.NotNil(t, lease)
		require.Equal(t, expectedLeaseDurationSeconds, lease.Spec.LeaseDurationSeconds)
		return true
	}, 5*time.Second, 100*time.Millisecond)

	require.NoError(t, lr.Shutdown(ctx))
}
