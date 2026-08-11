package theme_test

import (
	"testing"

	"keim/internal/theme"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestColorsAreDefined(t *testing.T) {
	tests := []struct {
		name  string
		color lipgloss.Color
	}{
		{"Primary", theme.Primary},
		{"Secondary", theme.Secondary},
		{"Accent", theme.Accent},
		{"Text", theme.Text},
		{"Muted", theme.Muted},
		{"Warning", theme.Warning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, string(tt.color))
		})
	}
}
