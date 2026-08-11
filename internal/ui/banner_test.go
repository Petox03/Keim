package ui_test

import (
	"bytes"
	"testing"

	"keim/internal/ui"

	"github.com/stretchr/testify/assert"
)

func TestPrintBanner(t *testing.T) {
	t.Run("Contains key banner elements", func(t *testing.T) {
		var buf bytes.Buffer

		err := ui.PrintBanner(&buf, "2.5.6")

		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "KEIM")
		assert.Contains(t, output, "v2.5.6")
		assert.Contains(t, output, "Scaffolder")
		assert.Contains(t, output, "Alberto Sosa")
	})

	t.Run("Error when writer is nil", func(t *testing.T) {
		err := ui.PrintBanner(nil, "2.5.6")

		assert.Error(t, err)
		assert.Equal(t, "el io.Writer no puede ser nil", err.Error())
	})

	t.Run("Version prefix normalization", func(t *testing.T) {
		tests := []struct {
			name    string
			input   string
			want    string
			notWant string
		}{
			{"hardcodeo sin v", "0.2.1", "v0.2.1", "vv0.2.1"},
			{"git describe con v", "v0.2.1", "v0.2.1", "vv0.2.1"},
			{"fallback dev", "dev", "dev", "vdev"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var buf bytes.Buffer
				err := ui.PrintBanner(&buf, tt.input)
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), tt.want)
				assert.NotContains(t, buf.String(), tt.notWant)
			})
		}
	})
}
