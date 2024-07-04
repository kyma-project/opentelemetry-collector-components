package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultComponents(t *testing.T) {
	factories, err := Components()
	require.NoError(t, err)

	exts := factories.Extensions
	for k, v := range exts {
		assert.Equal(t, k, v.Type())
	}

	recvs := factories.Receivers
	for k, v := range recvs {
		assert.Equal(t, k, v.Type())
	}

	procs := factories.Processors
	for k, v := range procs {
		assert.Equal(t, k, v.Type())
	}

	exps := factories.Exporters
	for k, v := range exps {
		assert.Equal(t, k, v.Type())
	}
}
