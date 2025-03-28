package serviceenrichmentprocessor

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor/processortest"

	"github.com/kyma-project/opentelemetry-collector-components/processor/serviceenrichmentprocessor/internal/metadata"
)

func TestProcessTraces(t *testing.T) {
	tt := []struct {
		name                string
		traces              ptrace.Traces
		expectedServiceName string
	}{
		{
			name: "traces with service name not set and k8s-io-app-name-set",
			traces: tracesWithResourceAttrs(map[string]string{
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "foo-k8s-io-app-name",
		},
		{
			name: "traces with service name not set and app-name-set",
			traces: tracesWithResourceAttrs(map[string]string{
				"kyma.app_name": "foo-app-name",
			}),
			expectedServiceName: "foo-app-name",
		},
		{
			name: "traces with service name not set and deployment name set",
			traces: tracesWithResourceAttrs(map[string]string{
				"k8s.deployment.name": "foo-deployment-name",
			}),
			expectedServiceName: "foo-deployment-name",
		},
		{
			name: "traces with service name not set and daemonset name set",
			traces: tracesWithResourceAttrs(map[string]string{
				"k8s.daemonset.name": "foo-daemonset-name",
			}),
			expectedServiceName: "foo-daemonset-name",
		},
		{
			name: "traces with service name not set and job name is set",
			traces: tracesWithResourceAttrs(map[string]string{
				"k8s.job.name": "foo-job-name",
			}),
			expectedServiceName: "foo-job-name",
		},
		{
			name: "traces with service name set to unknown_service",
			traces: tracesWithResourceAttrs(map[string]string{
				"service.name":                "unknown_service",
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "foo-k8s-io-app-name",
		},
		{
			name: "traces with service name set to test_unknown_service",
			traces: tracesWithResourceAttrs(map[string]string{
				"service.name":                "test_unknown_service",
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "test_unknown_service",
		},
		{
			name: "traces with service name set to unknown_service_test",
			traces: tracesWithResourceAttrs(map[string]string{
				"service.name":                "unknown_service_test",
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "unknown_service_test",
		},
		{
			name: "traces with service name set to unknown_service:",
			traces: tracesWithResourceAttrs(map[string]string{
				"service.name":                "unknown_service:",
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "unknown_service:",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sink := new(consumertest.TracesSink)

			config := Config{
				CustomLabels: []string{
					"kyma.kubernetes_io_app_name",
					"kyma.app_name",
				},
			}

			factory := NewFactory()
			cm, err := factory.CreateTraces(
				t.Context(),
				processortest.NewNopSettings(metadata.Type),
				config,
				sink,
			)
			require.NotNil(t, cm)
			require.NoError(t, err)

			err = cm.Start(t.Context(), componenttest.NewNopHost())
			require.NoError(t, err)

			cErr := cm.ConsumeTraces(t.Context(), tc.traces)
			require.NoError(t, cErr)

			got := sink.AllTraces()
			require.Len(t, got, 1)
			for _, tr := range got {
				for i := 0; i < tr.ResourceSpans().Len(); i++ {
					attr := tr.ResourceSpans().At(i).Resource().Attributes()
					svcName, ok := attr.Get("service.name")
					require.True(t, ok)
					require.Equal(t, tc.expectedServiceName, svcName.AsString())
				}
			}
		})
	}
}

func tracesWithResourceAttrs(attrs ...map[string]string) ptrace.Traces {
	traces := ptrace.NewTraces()
	for _, attr := range attrs {
		resTraces := traces.ResourceSpans().AppendEmpty()
		for k, v := range attr {
			resTraces.Resource().Attributes().PutStr(k, v)
		}
	}
	return traces
}
