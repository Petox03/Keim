package generator_test

import (
	"os"
	"path/filepath"
	"testing"

	"keim/internal/generator"
	"keim/internal/project"

	"github.com/stretchr/testify/assert"
)

func TestCreateProjectDir(t *testing.T) {
	tests := []struct {
		name      string
		getRoute  func(tmpDir string) string
		deepCheck bool
	}{
		{
			name: "Standard Path",
			getRoute: func(tmpDir string) string {
				return filepath.Join(tmpDir, "my-project")
			},
		},
		{
			name: "Deep Paths",
			getRoute: func(tmpDir string) string {
				return filepath.Join(tmpDir, "level1", "level2", "my-project")
			},
			deepCheck: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finalRoute := tt.getRoute(t.TempDir())

			err := generator.CreateProjectDir(finalRoute)
			assert.NoError(t, err)

			info, err := os.Stat(finalRoute)
			assert.NoError(t, err)
			assert.True(t, info.IsDir(), "Se esperaba que la ruta fuera un directorio")
		})
	}
}

func TestIdempotenceCallTwice(t *testing.T) {
	finalRoute := filepath.Join(t.TempDir(), "my-project")

	err := generator.CreateProjectDir(finalRoute)
	assert.NoError(t, err)

	witnessFile := filepath.Join(finalRoute, "test.txt")

	err = os.WriteFile(witnessFile, []byte("datos de prueba"), 0644)
	assert.NoError(t, err)

	err = generator.CreateProjectDir(finalRoute)
	assert.NoError(t, err)

	_, err = os.Stat(witnessFile)
	assert.NoError(t, err)

	data, err := os.ReadFile(witnessFile)
	assert.NoError(t, err)
	assert.Equal(t, "datos de prueba", string(data))
}

func TestWriteFiles(t *testing.T) {
	t.Run("CreateAllFiles", func(t *testing.T) {
		tmpDir := t.TempDir()

		p := project.Project{
			Name:      "scaffold_test",
			Path:      tmpDir,
			GoVersion: "1.26",
		}

		files := map[string][]byte{
			"go.mod":        []byte("module scaffold_test\n\ngo 1.26\n"),
			"main.go":       []byte("package main\n\nfunc main() {}\n"),
			"Dockerfile":    []byte("FROM golang:1.26-alpine\n"),
			"compose.yml":   []byte("services:\n  app:\n"),
			".gitignore":    []byte("*.exe\n"),
			".dockerignore": []byte(".git\n"),
		}

		err := generator.WriteFiles(p, files)
		assert.NoError(t, err)

		for relPath, expectedContent := range files {
			fullPath := filepath.Join(tmpDir, relPath)

			content, err := os.ReadFile(fullPath)
			assert.NoError(t, err, "El archivo %s debió haber sido creado por el generador", relPath)
			assert.Equal(t, expectedContent, content, "El contenido del archivo %s no coincide", relPath)
		}
	})

	t.Run("WithInvalidPath", func(t *testing.T) {
		// /dev/null es un archivo, no un directorio.
		// os.MkdirAll no puede crear subcarpetas bajo él y debe fallar.
		invalidPath := "/dev/null/invalid-path"

		p := project.Project{
			Name:      "scaffold_invalid",
			Path:      invalidPath,
			GoVersion: "1.26",
		}

		files := map[string][]byte{
			"go.mod": []byte("module scaffold_invalid\n"),
		}

		err := generator.WriteFiles(p, files)
		assert.Error(t, err)
	})

	t.Run("CreatesSubfoldersAutomatically", func(t *testing.T) {
		tmpDir := t.TempDir()

		p := project.Project{
			Name:      "testapp",
			Path:      tmpDir,
			GoVersion: "1.26",
		}

		files := map[string][]byte{
			".devcontainer/devcontainer.json": []byte(`{"name":"testapp"}`),
		}

		err := generator.WriteFiles(p, files)
		assert.NoError(t, err)

		devcontainerPath := filepath.Join(tmpDir, ".devcontainer", "devcontainer.json")
		content, err := os.ReadFile(devcontainerPath)
		assert.NoError(t, err, "devcontainer.json debió haber sido creado en .devcontainer/")
		assert.Equal(t, `{"name":"testapp"}`, string(content))
	})

	t.Run("EmptyMapWritesNothing", func(t *testing.T) {
		tmpDir := t.TempDir()

		p := project.Project{
			Name:      "testapp",
			Path:      tmpDir,
			GoVersion: "1.26",
		}

		err := generator.WriteFiles(p, map[string][]byte{})
		assert.NoError(t, err)

		entries, err := os.ReadDir(tmpDir)
		assert.NoError(t, err)
		assert.Empty(t, entries, "El directorio del proyecto debió quedar vacío")
	})
}
