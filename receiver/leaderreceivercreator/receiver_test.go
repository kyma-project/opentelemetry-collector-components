package leaderreceivercreator

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/receiver/receivertest"
	v1 "k8s.io/api/coordination/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

type mockComponent struct {
}

func (c *mockComponent) Start(_ context.Context, _ component.Host) error {
	fmt.Println("Starting subreceiver...")
	return nil
}

func (c *mockComponent) Shutdown(_ context.Context) error {
	fmt.Println("Shutting down subreceiver...")
	return nil
}

func TestMockReceiverCreator(t *testing.T) {
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
	lr.makeClient = func() (kubernetes.Interface, error) {
		return fake.NewSimpleClientset(&v1.Lease{ObjectMeta: metav1.ObjectMeta{Name: "my-foo-lease-1", Namespace: "default"}}), nil
	}
	fakeClient, err := lr.makeClient()
	require.NoError(t, err)

	ctx := context.TODO()
	go lr.Start(ctx, componenttest.NewNopHost())

	require.Eventually(t, func() bool {
		lease, err := fakeClient.CoordinationV1().Leases("default").Get(ctx, "my-foo-lease-1", metav1.GetOptions{})
		require.NoError(t, err)
		return lease != nil
	}, 5*time.Second, 100*time.Millisecond)

	require.NoError(t, lr.Shutdown(ctx))
}
