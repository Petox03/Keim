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
	forbiddenFiles := templates.GetForbiddenFiles()

	tests := []struct {
		CaseName          string
		PreExistingFiles  []string
		ExpectedErrSuffix string
	}{
		{
			CaseName:          "Clean & valid route",
			PreExistingFiles:  []string{"README.md"},
			ExpectedErrSuffix: "",
		},
		{
			CaseName:          "Conflict with critical Go files",
			PreExistingFiles:  []string{"go.mod", "main.go", ".dockerignore", "README.md"},
			ExpectedErrSuffix: "contiene archivos en conflicto: .dockerignore, go.mod, main.go",
		},
		{
			CaseName:          "Conflict files",
			PreExistingFiles:  []string{".gitignore", "compose.yml", "Dockerfile", "go.mod", "main.go"},
			ExpectedErrSuffix: "contiene archivos en conflicto: .gitignore, Dockerfile, compose.yml, go.mod, main.go",
		},
		{
			CaseName:          "A completely empty route is valid",
			PreExistingFiles:  []string{},
			ExpectedErrSuffix: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.CaseName, func(t *testing.T) {
			tmpDir := t.TempDir()

			for _, file := range tt.PreExistingFiles {
				filePath := filepath.Join(tmpDir, file)
				err := os.WriteFile(filePath, []byte("contenido"), 0644)
				assert.NoError(t, err)
			}

			err := validator.Validate(tmpDir, forbiddenFiles)

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