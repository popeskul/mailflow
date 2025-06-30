package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistry_Init(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "registry should be initialized with default collectors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, Registry)

			// Test that we can gather metrics
			metrics, err := Registry.Gather()
			assert.NoError(t, err)
			assert.NotEmpty(t, metrics)

			// Should have at least go collector metrics
			foundGoCollector := false

			for _, mf := range metrics {
				name := mf.GetName()
				if name == "go_goroutines" {
					foundGoCollector = true
				}
			}

			assert.True(t, foundGoCollector, "Go collector metrics should be present")
			// Should have at least some metrics from collectors
			assert.True(t, len(metrics) >= 2, "Should have at least some metrics from collectors")
		})
	}
}
