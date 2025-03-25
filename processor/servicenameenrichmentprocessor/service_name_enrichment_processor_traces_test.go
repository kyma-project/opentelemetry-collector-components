package servicenameenrichmentprocessor

import (
	"context"
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
			name: "traces with service name not set",
			traces: setTraces(map[string]string{
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "foo-k8s-io-app-name",
		},
		{
			name: "traces with service name set to unknown_service",
			traces: setTraces(map[string]string{
				"service.name":                "unknown_service",
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "foo-k8s-io-app-name",
		},
		{
			name: "traces with service name not set and deployment name set",
			traces: setTraces(map[string]string{
				"k8s.deployment.name": "foo-deployment-name",
			}),
			expectedServiceName: "foo-deployment-name",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sink := new(consumertest.TracesSink)

			//logger := zap.NewNop()
			config := Config{
				CustomLabels: []string{
					"kyma.kubernetes_io_app_name",
					"kyma.app_name",
				},
			}
			//sep := newServiceEnrichmentProcessor(logger, config)
			//res, err := sep.processTraces(context.TODO(), tc.traces)

			factory := NewFactory()
			cm, err := factory.CreateTraces(
				context.TODO(),
				processortest.NewNopSettingsWithType(metadata.Type),
				config,
				sink,
			)
			err = cm.Start(context.Background(), componenttest.NewNopHost())
			require.NotNil(t, cm)
			require.NoError(t, err)

			cErr := cm.ConsumeTraces(context.TODO(), tc.traces)
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

func setTraces(attrs ...map[string]string) ptrace.Traces {
	traces := ptrace.NewTraces()
	for _, attr := range attrs {
		resTraces := traces.ResourceSpans().AppendEmpty()
		for k, v := range attr {
			resTraces.Resource().Attributes().PutStr(k, v)
		}
	}
	return traces
}
