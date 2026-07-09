package generator

import (
	"os"
	"path/filepath"

	"keim/internal/project"
	"keim/internal/templates"
)

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