package templates

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"sync"
	"text/template"

	"keim/internal/project"
)

//go:embed files/*.tmpl
var templatesFS embed.FS // Esto significa Embed File System que es cuando incrustas múltiples archivos o carpetas enteras.

// getTemplates es la función memorizada a nivel de paquete.
// Reemplaza a init() y a cacheTemplates.
// La primera vez que alguien llame a Render(), ejecutará el parseo.
// A partir de la segunda vez, devolverá inmediatamente el puntero memorizado o el error.
var getTemplates = sync.OnceValues(func() (*template.Template, error) {
	tmpl, err := template.ParseFS(templatesFS, "files/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("error al parsear plantillas en embed.FS: %w", err)
	}
	return tmpl, nil
})

// Como el paquete template es el que maneja todo lo relacionado con las plantillas del proyecto, vi coherente
// que sea este el que tenga una función que las liste completamente, así puedo usarlo en los demás archivos
// que necesiten observar, manejar, o saber del listado de plantillas que tenemos.
func FileNames() []string {
	var names []string
	files, err := templatesFS.ReadDir("files")
	if err != nil {
		return nil
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".tmpl") {
			cleanName := strings.TrimSuffix(file.Name(), ".tmpl")
			names = append(names, cleanName)
		}
	}
	return names
}

func Render(name string, data project.Project) ([]byte, error) {
	// 1. Carga Lazy: Obtenemos el árbol de plantillas parseado.
	// Si es la 1.ª vez, procesa files/*.tmpl.
	// Si es la 2.ª+ vez, entrega el cache en nanosegundos.
	// Si el parseo falló, propagamos el error hacia el caller (main/CLI) sin panic.
	cacheTemplates, err := getTemplates()
	if err != nil {
		return nil, err
	}

	originalName := name + ".tmpl"

	// 2. Preparar el buffer en memoria
	var templateResult bytes.Buffer

	// 3. Renderizar usando la colección en cache
	err = cacheTemplates.ExecuteTemplate(&templateResult, originalName, data)
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar la plantilla %q: %w", originalName, err)
	}

	// 4. Retornar los bytes generados
	return templateResult.Bytes(), nil
}