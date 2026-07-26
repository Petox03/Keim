package ui_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"keim/internal/project"
	"keim/internal/ui"
)

func TestPrintReport(t *testing.T) {
	files := []string{".dockerignore", ".gitignore", "Dockerfile", "compose.yml", "go.mod", "main.go"}

	tests := []struct {
		CaseName      string
		ProjectPath   string
		ExpectedSteps []string
		UnwantedStep  string
	}{
		{
			CaseName:    "Project in subdirectory includes cd command",
			ProjectPath: "clippy",
			ExpectedSteps: []string{
				"1. cd clippy",
				"2. docker compose up -d",
				"3. docker compose exec app go run .",
			},
			UnwantedStep: "",
		},
		{
			CaseName:    "Project in current directory omits cd command",
			ProjectPath: ".",
			ExpectedSteps: []string{
				"1. docker compose up -d",
				"2. docker compose exec app go run .",
			},
			UnwantedStep: "cd .",
		},
	}

	for _, tt := range tests {
		t.Run(tt.CaseName, func(t *testing.T) {
			var buf bytes.Buffer
			p := project.Project{
				Name:      "clippy",
				Path:      tt.ProjectPath,
				GoVersion: "1.25",
			}

			err := ui.PrintReport(&buf, p, files)

			assert.NoError(t, err)

			output := buf.String()
			assert.Contains(t, output, "Proyecto 'clippy' creado con éxito!")

			for _, step := range tt.ExpectedSteps {
				assert.Contains(t, output, step)
			}

			if tt.UnwantedStep != "" {
				assert.NotContains(t, output, tt.UnwantedStep)
			}
		})
	}

	t.Run("Error: Writer is nil", func(t *testing.T) {
		p := project.Project{Name: "clippy", Path: "clippy", GoVersion: "1.25"}

		err := ui.PrintReport(nil, p, files)

		assert.Error(t, err)
		assert.Equal(t, "el io.Writer no puede ser nil", err.Error())
	})
}
