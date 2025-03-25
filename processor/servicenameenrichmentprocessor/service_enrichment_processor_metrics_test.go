package servicenameenrichmentprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor/processortest"

	"github.com/kyma-project/opentelemetry-collector-components/processor/serviceenrichmentprocessor/internal/metadata"
)

func TestProcessMetrics(t *testing.T) {
	tt := []struct {
		name                string
		metrics             pmetric.Metrics
		expectedServiceName string
	}{
		{
			name: "metrics with service name not set",
			metrics: setMetrics(map[string]string{
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "foo-k8s-io-app-name",
		},
		{
			name: "metrics with service name set to unknown_service",
			metrics: setMetrics(map[string]string{
				"service.name":                "unknown_service",
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "foo-k8s-io-app-name",
		},
		{
			name: "metrics with service name not set and deployment name set",
			metrics: setMetrics(map[string]string{
				"k8s.deployment.name": "foo-deployment-name",
			}),
			expectedServiceName: "foo-deployment-name",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			//logger := zap.NewNop()
			config := Config{
				CustomLabels: []string{
					"kyma.kubernetes_io_app_name",
					"kyma.app_name",
				},
			}
			sink := new(consumertest.MetricsSink)
			factory := NewFactory()
			cm, err := factory.CreateMetrics(
				context.TODO(),
				processortest.NewNopSettingsWithType(metadata.Type),
				config,
				sink,
			)
			err = cm.Start(context.Background(), componenttest.NewNopHost())
			require.NotNil(t, cm)
			require.NoError(t, err)

			cErr := cm.ConsumeMetrics(context.TODO(), tc.metrics)
			require.NoError(t, cErr)

			got := sink.AllMetrics()
			require.Len(t, got, 1)

			for _, m := range got {
				for i := 0; i < m.ResourceMetrics().Len(); i++ {
					attr := m.ResourceMetrics().At(i).Resource().Attributes()
					svcName, ok := attr.Get("service.name")
					require.True(t, ok)
					require.Equal(t, tc.expectedServiceName, svcName.AsString())
				}
			}
		})
	}
}

func setMetrics(attrs ...map[string]string) pmetric.Metrics {
	metrics := pmetric.NewMetrics()
	for _, attr := range attrs {
		resMetrics := metrics.ResourceMetrics().AppendEmpty()
		for k, v := range attr {
			resMetrics.Resource().Attributes().PutStr(k, v)
		}
	}
	return metrics
}
