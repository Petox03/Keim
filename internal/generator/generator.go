package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"keim/internal/project"
	"keim/internal/templates"
)

func CreateProjectDir(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("error al crear el directorio del proyecto '%s': %w", path, err)
	}
	return nil
}

func Generate(p project.Project, filesToGenerate []string) error {

	for _, fileName := range filesToGenerate {
		renderedBytes, err := templates.Render(fileName, p)
		if err != nil {
			return err
		}

		path := filepath.Join(p.Path, fileName)
		err = os.WriteFile(path, renderedBytes, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}