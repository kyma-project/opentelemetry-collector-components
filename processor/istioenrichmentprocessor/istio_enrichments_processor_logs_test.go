package istioenrichmentprocessor

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/processor/processortest"

	"github.com/kyma-project/opentelemetry-collector-components/processor/istioenrichmentprocessor/internal/metadata"
)

func TestProcessLogs(t *testing.T) {
	tt := []struct {
		name                               string
		logs                               plog.Logs
		expectedScopeName                  string
		expectedScopeVersion               string
		expectedNetworkProtocolName        string
		expectedNetworkProtocolVersion     string
		expectedClientAddress              string
		expectedClientPort                 string
		expectedSeverityText               string
		expectedSeverityNumber             plog.SeverityNumber
		expectShouldTestModifiedAttributes bool
	}{
		{
			name: "logs with kyma.module istio attribute",
			logs: NewPLogBuilder().WithScopeName("test_scope").
				WithScopeVersion("someVersion").
				WithLogAttributes(map[string]string{
					"kyma.module":           "istio",
					"network.protocol.name": "HTTP/1.0",
					"server.address":        "server.local:123",
					"client.address":        "client.local:456",
				}).
				Build(),
			expectedScopeName:                  istioScopeName,
			expectedScopeVersion:               "v1",
			expectedNetworkProtocolName:        "HTTP",
			expectedNetworkProtocolVersion:     "1.0",
			expectedClientAddress:              "client.local",
			expectedClientPort:                 "456",
			expectedSeverityText:               "INFO",
			expectedSeverityNumber:             plog.SeverityNumberInfo,
			expectShouldTestModifiedAttributes: true,
		},
		{
			name: "logs without kyma.module istio attribute",
			logs: NewPLogBuilder().WithScopeName("test_scope").
				WithScopeVersion("someVersion").
				WithLogAttributes(map[string]string{
					"kyma.module":           "some_other_module",
					"network.protocol.name": "HTTP/1.0",
					"server.address":        "server.local:123",
					"client.address":        "client.local:456",
				}).
				Build(),
			expectedScopeName:                  "test_scope",
			expectedScopeVersion:               "someVersion",
			expectedNetworkProtocolName:        "HTTP/1.0",
			expectedClientAddress:              "client.local:456",
			expectedSeverityText:               "",
			expectedSeverityNumber:             plog.SeverityNumberUnspecified,
			expectShouldTestModifiedAttributes: false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sink := new(consumertest.LogsSink)
			config := Config{
				ScopeVersion: "v1",
			}

			factory := NewFactory()
			cm, err := factory.CreateLogs(
				t.Context(),
				processortest.NewNopSettings(metadata.Type),
				config,
				sink,
			)
			require.NoError(t, err)
			require.NotNil(t, cm)

			err = cm.Start(t.Context(), componenttest.NewNopHost())
			require.NoError(t, err)

			cErr := cm.ConsumeLogs(t.Context(), tc.logs)
			require.NoError(t, cErr)

			allLogs := sink.AllLogs()
			require.Len(t, allLogs, 1)
			for _, l := range allLogs {
				for _, r := range l.ResourceLogs().All() {
					for _, s := range r.ScopeLogs().All() {

						require.Equal(t, tc.expectedScopeVersion, s.Scope().Version())
						require.Equal(t, tc.expectedScopeName, s.Scope().Name())

						for _, lr := range s.LogRecords().All() {
							require.Equal(t, tc.expectedNetworkProtocolName, lr.Attributes().AsRaw()["network.protocol.name"])
							require.Equal(t, tc.expectedClientAddress, lr.Attributes().AsRaw()["client.address"])
							require.Equal(t, tc.expectedSeverityNumber, lr.SeverityNumber())
							require.Equal(t, tc.expectedSeverityText, lr.SeverityText())

							if tc.expectShouldTestModifiedAttributes {
								require.Equal(t, tc.expectedNetworkProtocolVersion, lr.Attributes().AsRaw()["network.protocol.version"])
								require.Equal(t, tc.expectedClientPort, lr.Attributes().AsRaw()["client.port"])
							}
						}
					}
				}
			}
		})
	}
}
