package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"keim/internal/project"
	"keim/internal/templates"
)

// CreateProjectDir crea la estructura de directorios necesaria.
func CreateProjectDir(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("error al crear el directorio del proyecto '%s': %w", path, err)
	}
	return nil
}

// Generate renderiza todas las plantillas en memoria primero.
// Solo si el 100% de las plantillas se procesan con éxito, escribe los archivos a disco.
func Generate(p project.Project, filesToGenerate []string) error {
	renderedFiles := make(map[string][]byte, len(filesToGenerate))

	// Fase 1: Renderizar TODO en memoria RAM.
	// Si un template tiene un error de sintaxis o falta un dato, abortamos sin tocar el disco.
	for _, fileName := range filesToGenerate {
		renderedBytes, err := templates.Render(fileName, p)
		if err != nil {
			return err
		}
		renderedFiles[fileName] = renderedBytes
	}

	// Fase 2: Persistencia a disco.
	// Garantizado que todo el contenido en memoria es válido.
	for _, fileName := range filesToGenerate {
		path := filepath.Join(p.Path, fileName)
		err := os.WriteFile(path, renderedFiles[fileName], 0644)
		if err != nil {
			return fmt.Errorf("error al escribir el archivo '%s': %w", path, err)
		}
	}

	return nil
}