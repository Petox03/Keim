package validator_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"keim/internal/templates"
	"keim/internal/validator"
)

func TestValidate(t *testing.T) {
	forbiddenFiles := templates.GetForbiddenFiles(false)
	forbiddenFilesWithDevcontainer := templates.GetForbiddenFiles(true)

	tests := []struct {
		CaseName          string
		Forbidden         map[string]bool
		PreExistingFiles  []string
		ExpectedErrSuffix string
	}{
		{
			CaseName:          "Clean & valid route",
			Forbidden:         forbiddenFiles,
			PreExistingFiles:  []string{"README.md"},
			ExpectedErrSuffix: "",
		},
		{
			CaseName:          "Conflict with critical Go files",
			Forbidden:         forbiddenFiles,
			PreExistingFiles:  []string{"go.mod", "main.go", ".dockerignore", "README.md"},
			ExpectedErrSuffix: "contiene archivos en conflicto: .dockerignore, go.mod, main.go",
		},
		{
			CaseName:          "Conflict files",
			Forbidden:         forbiddenFiles,
			PreExistingFiles:  []string{".gitignore", "compose.yml", "Dockerfile", "go.mod", "main.go"},
			ExpectedErrSuffix: "contiene archivos en conflicto: .gitignore, Dockerfile, compose.yml, go.mod, main.go",
		},
		{
			CaseName:          "A completely empty route is valid",
			Forbidden:         forbiddenFiles,
			PreExistingFiles:  []string{},
			ExpectedErrSuffix: "",
		},
		{
			// Bug original: devcontainer.json en .devcontainer/ no se detectaba
			// porque el validator sólo leía la raíz del proyecto.
			CaseName:          "Conflict with devcontainer.json in .devcontainer subfolder",
			Forbidden:         forbiddenFilesWithDevcontainer,
			PreExistingFiles:  []string{".devcontainer/devcontainer.json"},
			ExpectedErrSuffix: "contiene archivos en conflicto: .devcontainer/devcontainer.json",
		},
		{
			// Falso positivo original: devcontainer.json en la raíz se marcaba
			// como conflicto aunque Keim lo genera en .devcontainer/, no en la raíz.
			CaseName:          "devcontainer.json in root is NOT a conflict (Keim generates it in .devcontainer/)",
			Forbidden:         forbiddenFilesWithDevcontainer,
			PreExistingFiles:  []string{"devcontainer.json"},
			ExpectedErrSuffix: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.CaseName, func(t *testing.T) {
			tmpDir := t.TempDir()

			for _, file := range tt.PreExistingFiles {
				filePath := filepath.Join(tmpDir, file)
				err := os.MkdirAll(filepath.Dir(filePath), 0755)
				assert.NoError(t, err)
				err = os.WriteFile(filePath, []byte("contenido"), 0644)
				assert.NoError(t, err)
			}

			err := validator.Validate(tmpDir, tt.Forbidden)

			if tt.ExpectedErrSuffix == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.ExpectedErrSuffix)
			}
		})
	}

	t.Run("Error: Route does not exist", func(t *testing.T) {
		fakePath := "./ruta/completamente/inexistente/falsa"
		err := validator.Validate(fakePath, forbiddenFiles)

		assert.ErrorIs(t, err, validator.ErrPathNotFound)
	})

	t.Run("Error: Route is a file, not a directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "un_archivo.txt")

		err := os.WriteFile(filePath, []byte("hola"), 0644)
		assert.NoError(t, err)

		err = validator.Validate(filePath, forbiddenFiles)

		assert.Error(t, err)
		assert.ErrorContains(t, err, fmt.Sprintf("la ruta '%s' no es accesible:", filePath))
	})
}
