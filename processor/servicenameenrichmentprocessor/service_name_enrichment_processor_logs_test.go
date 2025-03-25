package servicenameenrichmentprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/processor/processortest"

	"github.com/kyma-project/opentelemetry-collector-components/processor/serviceenrichmentprocessor/internal/metadata"
)

func TestProcessLogs(t *testing.T) {
	tt := []struct {
		name                string
		logs                plog.Logs
		expectedServiceName string
	}{
		{
			name: "logs with service name not set",
			logs: setLogs(map[string]string{
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "foo-k8s-io-app-name",
		},
		{
			name: "logs with service name set to unknown_service",
			logs: setLogs(map[string]string{
				"service.name":                "unknown_service",
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "foo-k8s-io-app-name",
		},
		{
			name: "logs with service name not set and deployment name set",
			logs: setLogs(map[string]string{
				"k8s.deployment.name": "foo-deployment-name",
			}),
			expectedServiceName: "foo-deployment-name",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sink := new(consumertest.LogsSink)
			//logger := zap.NewNop()
			config := Config{
				CustomLabels: []string{
					"kyma.kubernetes_io_app_name",
					"kyma.app_name",
				},
			}
			//sep := newServiceEnrichmentProcessor(logger, config)
			//res, err := sep.processLogs(context.TODO(), tc.logs)
			//require.NoError(t, err)

			factory := NewFactory()
			cm, err := factory.CreateLogs(
				context.TODO(),
				processortest.NewNopSettingsWithType(metadata.Type),
				config,
				sink,
			)
			err = cm.Start(context.Background(), componenttest.NewNopHost())
			require.NotNil(t, cm)
			require.NoError(t, err)

			cErr := cm.ConsumeLogs(context.TODO(), tc.logs)
			require.NoError(t, cErr)

			got := sink.AllLogs()
			require.Len(t, got, 1)
			for _, l := range got {
				for i := 0; i < l.ResourceLogs().Len(); i++ {
					attr := l.ResourceLogs().At(i).Resource().Attributes()
					svcName, ok := attr.Get("service.name")
					require.True(t, ok)
					require.Equal(t, tc.expectedServiceName, svcName.AsString())
				}
			}
		})
	}
}

func setLogs(attrs ...map[string]string) plog.Logs {
	logs := plog.NewLogs()
	for _, attr := range attrs {
		resLogs := logs.ResourceLogs().AppendEmpty()
		for k, v := range attr {
			resLogs.Resource().Attributes().PutStr(k, v)
		}
	}
	return logs
}
