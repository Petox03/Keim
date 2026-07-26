package ui

import (
	"fmt"
	"io"
	"path/filepath"

	"keim/internal/project"
)

func PrintReport(w io.Writer, p project.Project, files []string) error {

	if w == nil {
		return fmt.Errorf("el io.Writer no puede ser nil")
	}

	// 1. Encabezado principal y detalles del proyecto
	_, err := fmt.Fprintf(
		w,
		"Proyecto '%s' creado con éxito!\n\n📌 Detalles del proyecto:\n	• Ruta:           %s\n	• Versión de Go:  %s\n",
		p.Name,
		p.Path,
		p.GoVersion,
	)
	if err != nil {
		return fmt.Errorf("falló la escritura en io.Writer: %w", err)
	}

	// 2. Sección de archivos generados
	_, err = fmt.Fprintln(w, "\nArchivos generados:")
	if err != nil {
		return fmt.Errorf("error al escribir encabezado de archivos: %w", err)
	}

	for _, file := range files {
		_, err := fmt.Fprintf(w, "	• %s\n", file)
		if err != nil {
			return fmt.Errorf("error al escribir el archivo %q: %w", file, err)
		}
	}

	// 3. Sección de siguientes pasos
	_, err = fmt.Fprintln(w, "\nSiguientes pasos:")
	if err != nil {
		return fmt.Errorf("error al escribir encabezado de pasos: %w", err)
	}

	cleanPath := filepath.Clean(p.Path)

	// Si el proyecto se creó en el directorio actual, omitimos el paso "cd"
	if cleanPath == "." || cleanPath == "" {
		const tpl = `	1. docker compose up -d
	2. docker compose exec app go run .
`
		_, err := fmt.Fprint(w, tpl)
		if err != nil {
			return fmt.Errorf("error al escribir pasos: %w", err)
		}
	} else {
		const tpl = `	1. cd %s
	2. docker compose up -d
	3. docker compose exec app go run .
`
		_, err := fmt.Fprintf(w, tpl, cleanPath)
		if err != nil {
			return fmt.Errorf("error al escribir pasos: %w", err)
		}
	}

	return nil
}
