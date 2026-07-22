package validator

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

var ErrPathNotFound = errors.New("la ruta especificada no existe")

// Validate comprueba si un directorio es apto para trabajar.
// Retorna nil si está limpio, o un error descriptivo (con la ruta) si no existe,
// no es accesible o contiene archivos en conflicto.
func Validate(path string, keyFiles []string) error {
	forbiddenFiles := make(map[string]bool)

	for _, file := range keyFiles {
		lower := strings.ToLower(file)
		forbiddenFiles[lower] = true

		if lower == "compose.yml" || lower == "compose.yaml" {
			forbiddenFiles["compose.yml"] = true
			forbiddenFiles["compose.yaml"] = true
			forbiddenFiles["docker-compose.yml"] = true
			forbiddenFiles["docker-compose.yaml"] = true
		}
	}

	// 1. Intentamos leer el directorio directamente
	files, err := os.ReadDir(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%w: '%s'", ErrPathNotFound, path)
		}
		// Cubre permisos denegados, path es un archivo plano, etc.
		// Envolvemos el error original para preservar información de diagnóstico.
		return fmt.Errorf("la ruta '%s' no es accesible: %w", path, err)
	}

	var conflicts []string

	// 2. Iteramos sobre los elementos encontrados
	for _, file := range files {
		// Asegurarnos de que sea un archivo y no una subcarpeta con el mismo nombre
		if !file.IsDir() {
			fileName := strings.ToLower(file.Name())
			if forbiddenFiles[fileName] {
				conflicts = append(conflicts, file.Name())
			}
		}
	}

	// 3. Si hay conflictos, los unimos en un único error descriptivo
	if len(conflicts) > 0 {
		return fmt.Errorf("la ruta '%s' contiene archivos en conflicto: %s", path, strings.Join(conflicts, ", "))
	}

	return nil
}