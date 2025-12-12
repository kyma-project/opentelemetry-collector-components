package serviceenrichmentprocessor

import (
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
			name: "metrics with service name not set and k8s-io-app-name set",
			metrics: metricsWithResourceAttrs(map[string]string{
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "foo-k8s-io-app-name",
		},
		{
			name: "metrics with service name not set and app-name set",
			metrics: metricsWithResourceAttrs(map[string]string{
				"kyma.app_name": "foo-app-name",
			}),
			expectedServiceName: "foo-app-name",
		},
		{
			name: "metrics with service name not set and deployment name set",
			metrics: metricsWithResourceAttrs(map[string]string{
				"k8s.deployment.name": "foo-deployment-name",
			}),
			expectedServiceName: "foo-deployment-name",
		},
		{
			name: "metrics with service name not set and daemonset name set",
			metrics: metricsWithResourceAttrs(map[string]string{
				"k8s.daemonset.name": "foo-daemonset-name",
			}),
			expectedServiceName: "foo-daemonset-name",
		},
		{
			name: "metrics with service name not set and job name is set",
			metrics: metricsWithResourceAttrs(map[string]string{
				"k8s.job.name": "foo-job-name",
			}),
			expectedServiceName: "foo-job-name",
		},
		{
			name: "metrics with service name set to unknown_service",
			metrics: metricsWithResourceAttrs(map[string]string{
				"service.name":                "unknown_service",
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "foo-k8s-io-app-name",
		},
		{
			name: "metrics with service name set to test_unknown_service",
			metrics: metricsWithResourceAttrs(map[string]string{
				"service.name":                "test_unknown_service",
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "test_unknown_service",
		},
		{
			name: "metrics with service name set to unknown_service_test",
			metrics: metricsWithResourceAttrs(map[string]string{
				"service.name":                "unknown_service_test",
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "unknown_service_test",
		},
		{
			name: "metrics with service name set to unknown_service:",
			metrics: metricsWithResourceAttrs(map[string]string{
				"service.name":                "unknown_service:",
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "unknown_service:",
		},
		{
			name: "metrics with service name set to unknown_service:",
			metrics: metricsWithResourceAttrs(map[string]string{
				"service.name": "unknown_service:java",
			}),
			expectedServiceName: "unknown_service:java",
		},
		{
			name: "metrics with empty service name set and k8s-io-app-name set",
			metrics: metricsWithResourceAttrs(map[string]string{
				"service.name": "",
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "foo-k8s-io-app-name",
		},
		{
			name: "metrics with empty service name set",
			metrics: metricsWithResourceAttrs(map[string]string{
				"service.name": "",
			}),
			expectedServiceName: "unknown_service",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			config := Config{
				ResourceAttributes: []string{
					"kyma.kubernetes_io_app_name",
					"kyma.app_name",
				},
			}
			sink := new(consumertest.MetricsSink)
			factory := NewFactory()
			cm, err := factory.CreateMetrics(
				t.Context(),
				processortest.NewNopSettings(metadata.Type),
				config,
				sink,
			)
			require.NoError(t, err)
			require.NotNil(t, cm)
			err = cm.Start(t.Context(), componenttest.NewNopHost())
			require.NoError(t, err)

			cErr := cm.ConsumeMetrics(t.Context(), tc.metrics)
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

func metricsWithResourceAttrs(attrs ...map[string]string) pmetric.Metrics {
	metrics := pmetric.NewMetrics()
	for _, attr := range attrs {
		resMetrics := metrics.ResourceMetrics().AppendEmpty()
		for k, v := range attr {
			resMetrics.Resource().Attributes().PutStr(k, v)
		}
	}

	return metrics
}
