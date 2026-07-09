package generator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"keim/internal/generator"
	"keim/internal/project"
	"keim/internal/templates"
)

var expectedFiles = templates.FileNames()

func TestGenerateCreateAllFiles(t *testing.T) {
	// Crear directorio temporal donde el generador pueda escribir
	tmpDir := t.TempDir()

	p := project.Project{
		Name: 		"scaffold_test",
		Path: 		tmpDir,
		GoVersion:	"1.26",
	}

	// Ejecutar generador
	err := generator.Generate(p, expectedFiles)
	assert.NoError(t, err)

	// Obtener dinámicamente qué archivos se supone que debió haber creado
	assert.NotEmpty(t, expectedFiles,"La caché de templates no debería estar vacía")

	// Verificar en el disco que cada archivo exista en el tmpDir
	for _, fileName := range expectedFiles {
		path := filepath.Join(tmpDir, fileName)

		_, err := os.Stat(path)
		assert.NoError(t, err, "El archivo %s debió haber sido creado por el generador", fileName)

		// Opcional pero validamos que no están vacíos
		content, err := os.ReadFile(path)
		assert.NoError(t, err)
		assert.NotEmpty(t, content, "El archivo %s se creó vacío", fileName)
	}
}

func TestGenerateWithInvalidPath(t *testing.T) {
	invalidPath := "./invalid-path-totalmente-falso"

	p := project.Project{
		Name:      "scaffold_invalid",
		Path:      invalidPath,
		GoVersion: "1.26",
	}

	err := generator.Generate(p, expectedFiles) // asumiendo que ya borraste el expectedFiles de los parámetros
	assert.Error(t, err)

	// --- LA CORONACIÓN DEL TEST ---
	// Aseguramos que la carpeta ni siquiera se creó en el disco duro.
	_, statErr := os.Stat(invalidPath)
	assert.True(t, os.IsNotExist(statErr), "La ruta inválida no debió haber sido creada en el disco")

	// Limpieza por si las moscas (buenas prácticas de testing de I/O)
	defer os.RemoveAll(invalidPath)
}