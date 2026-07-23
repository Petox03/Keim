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
	// Obtenemos la lista negra O(1) de archivos en conflicto desde templates
	forbiddenFiles := templates.GetForbiddenFiles()

	tests := []struct {
		CaseName          string
		PreExistingFiles  []string
		ExpectedErrSubstring string // Subcadena esperada en el error ("" si debe pasar)
	}{
		{
			CaseName:             "Clean & valid route",
			PreExistingFiles:     []string{"README.md"},
			ExpectedErrSubstring: "",
		},
		{
			CaseName:             "Conflict with dockerfile",
			PreExistingFiles:     []string{"dockerfile"},
			ExpectedErrSubstring: "conflicto detectado: el archivo \"dockerfile\" ya existe",
		},
		{
			CaseName:             "Conflict with compose variant",
			PreExistingFiles:     []string{"compose.yaml"},
			ExpectedErrSubstring: "conflicto detectado: el archivo \"compose.yaml\" ya existe",
		},
		{
			CaseName:             "A completely empty route is valid",
			PreExistingFiles:     []string{},
			ExpectedErrSubstring: "",
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

			if tt.ExpectedErrSubstring == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.ExpectedErrSubstring)
			}
		})
	}

	// --- SUBTESTS AISLADOS PARA ESCENARIOS DE ERROR DEL PAQUETE OS ---

	t.Run("Error: Route does not exist", func(t *testing.T) {
		fakePath := filepath.Join(t.TempDir(), "ruta_inexistente")
		err := validator.Validate(fakePath, forbiddenFiles)

		assert.ErrorIs(t, err, validator.ErrPathNotFound)
	})

	t.Run("Error: Route is a file, not a directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "un_archivo.txt")

		err := os.WriteFile(filePath, []byte("hola"), 0644)
		assert.NoError(t, err)

		// Evaluamos pasándole un archivo en lugar del directorio
		err = validator.Validate(filePath, forbiddenFiles)

		assert.Error(t, err)
		assert.ErrorContains(t, err, fmt.Sprintf("no se pudo leer el directorio %s:", filePath))
	})
}