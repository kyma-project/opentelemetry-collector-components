package istionoisefilter

import (
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/ptracetest"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor/processortest"

	"github.com/kyma-project/opentelemetry-collector-components/processor/istionoisefilter/internal/metadata"
)

func TestIstioNoiseFilter(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()

	tp, err := factory.CreateTraces(t.Context(), processortest.NewNopSettings(metadata.Type), cfg, nil)
	require.NoError(t, err)
	require.NotNil(t, tp)

	td := ptrace.NewTraces()
	err = tp.ConsumeTraces(t.Context(), td)
	require.NoError(t, err)
	require.NoError(t, ptracetest.CompareTraces(td, ptrace.NewTraces()))
}
