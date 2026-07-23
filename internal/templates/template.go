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
var templatesFS embed.FS

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

// conflictRules define explícitamente qué variantes de nombres chocan entre sí.
// Si agregas una plantilla nueva, solo debes declarar sus variantes aquí.
var conflictRules = map[string][]string{
	"compose": {
		"compose.yml",
		"compose.yaml",
		"docker-compose.yml",
		"docker-compose.yaml",
	},
	"dockerfile": {
		"dockerfile",
		"dockerfile.dev",
		"dockerfile.prod",
	},
}

// GetForbiddenFiles transforma la tabla de reglas en un conjunto (set)
// de búsqueda rápida O(1) en minúsculas para el validador.
func GetForbiddenFiles() map[string]bool {
	forbidden := make(map[string]bool)

	// 1. Por defecto, TODOS los archivos que Keim genera son prohibidos si ya existen
    for _, name := range FileNames() {
        forbidden[strings.ToLower(name)] = true
    }

	// 2. Agregamos las variantes/alias explícitos definidos en las reglas
	for _, files := range conflictRules {
		for _, file := range files {
			forbidden[strings.ToLower(file)] = true
		}
	}

	return forbidden
}

// FileNames inspecciona directamente embed.FS para saber qué archivos genera Keim.
// Es ultra rápido porque solo lee el directorio virtual en memoria sin parsear nada.
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

// Render procesa bajo demanda la plantilla solicitada e inyecta los datos.
func Render(name string, data project.Project) ([]byte, error) {
	// 1. Carga Lazy: Obtenemos el árbol de plantillas parseado.
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