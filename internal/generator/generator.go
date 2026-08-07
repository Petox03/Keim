package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"keim/internal/project"
)

// CreateProjectDir crea la estructura de directorios necesaria.
func CreateProjectDir(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("error al crear el directorio del proyecto '%s': %w", path, err)
	}
	return nil
}

// WriteFiles escribe los archivos renderizados a disco.
// La clave del mapa es la ruta relativa del archivo (ej. "Dockerfile", ".devcontainer/devcontainer.json").
// Las subcarpetas necesarias se crean automáticamente.
func WriteFiles(p project.Project, files map[string][]byte) error {
	for relPath, content := range files {
		fullPath := filepath.Join(p.Path, relPath)

		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error al crear la subcarpeta '%s': %w", dir, err)
		}

		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			return fmt.Errorf("error al escribir el archivo '%s': %w", fullPath, err)
		}
	}

	return nil
}
