package validator

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

var ErrPathNotFound = errors.New("la ruta no existe")

// Validate comprueba que en el directorio destino no existan archivos en conflicto.
// Es completamente agnóstico a las plantillas de Keim; solo consume la lista negra recibida.
func Validate(targetDir string, forbiddenFiles map[string]bool) error {
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrPathNotFound
		}
		return fmt.Errorf("no se pudo leer el directorio %s: %w", targetDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileNameLower := strings.ToLower(entry.Name())

		// Búsqueda O(1) en el map[string]bool
		if forbiddenFiles[fileNameLower] {
			return fmt.Errorf("conflicto detectado: el archivo %q ya existe en el destino", entry.Name())
		}
	}

	return nil
}