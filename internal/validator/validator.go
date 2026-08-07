package validator

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var ErrPathNotFound = errors.New("la ruta especificada no existe")

// Validate comprueba si un directorio es apto para trabajar.
// Recibe la lista negra de archivos en conflicto (map[string]bool) y es completamente
// agnóstico a las plantillas internas de Keim.
// La clave del mapa es la ruta relativa del archivo (ej. "Dockerfile", ".devcontainer/devcontainer.json").
func Validate(path string, forbiddenFiles map[string]bool) error {
	// 1. Verificar que la ruta existe y es un directorio.
	// Si no existe, el caller (main.go) la creará. Si es un archivo, es un error fatal.
	if _, err := os.ReadDir(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%w: '%s'", ErrPathNotFound, path)
		}
		return fmt.Errorf("la ruta '%s' no es accesible: %w", path, err)
	}

	// 2. Verificar si alguna ruta prohibida ya existe en disco (case-insensitive).
	var conflicts []string

	for relPath := range forbiddenFiles {
		parentDir := filepath.Join(path, filepath.Dir(relPath))
		baseName := strings.ToLower(filepath.Base(relPath))

		entries, err := os.ReadDir(parentDir)
		if err != nil {
			// El directorio padre no existe, no hay conflicto posible para esta ruta.
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() && strings.ToLower(entry.Name()) == baseName {
				conflicts = append(conflicts, filepath.Join(filepath.Dir(relPath), entry.Name()))
				break
			}
		}
	}

	if len(conflicts) > 0 {
		sort.Strings(conflicts)
		return fmt.Errorf("la ruta '%s' contiene archivos en conflicto: %s", path, strings.Join(conflicts, ", "))
	}

	return nil
}
