package templates

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"keim/internal/project"
)

//go:embed files/*.tmpl files/optional/*.tmpl
var templatesFS embed.FS

// getTemplates es la función memorizada a nivel de paquete.
// Reemplaza a init() y a cacheTemplates.
// La primera vez que alguien llame a Render(), ejecutará el parseo.
// A partir de la segunda vez, devolverá inmediatamente el puntero memorizado o el error.
var getTemplates = sync.OnceValues(func() (*template.Template, error) {
	tmpl, err := template.ParseFS(templatesFS, "files/*.tmpl", "files/optional/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("error al parsear plantillas en embed.FS: %w", err)
	}
	return tmpl, nil
})

type FileSpec struct {
	Name string
	Dir  string
}

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

// targetOverrides mapea los archivos que van a una subcarpeta específica.
var targetOverrides = map[string]string{
	"devcontainer.json": ".devcontainer",
}

// GetForbiddenFiles transforma la tabla de reglas en un conjunto (set)
// de búsqueda rápida O(1) en minúsculas para el validador.
// La clave es la ruta relativa del archivo (ej. "Dockerfile", ".devcontainer/devcontainer.json").
func GetForbiddenFiles(withDevcontainer bool) map[string]bool {
	forbidden := make(map[string]bool)

	// 1. Por defecto, TODOS los archivos que Keim genera son prohibidos si ya existen
	for _, spec := range AllFileNames(withDevcontainer) {
		relPath := spec.Name
		if spec.Dir != "" {
			relPath = filepath.Join(spec.Dir, spec.Name)
		}
		forbidden[strings.ToLower(relPath)] = true
	}

	// 2. Agregamos las variantes/alias explícitos definidos en las reglas
	for _, files := range conflictRules {
		for _, file := range files {
			forbidden[strings.ToLower(file)] = true
		}
	}

	return forbidden
}

func AllFileNames(withDevcontainer bool) []FileSpec {
	names := FileNames()
	names = append(names, OptionalFileNames(withDevcontainer)...)
	return names
}

// FileNames inspecciona directamente embed.FS para saber qué archivos genera Keim.
func FileNames() []FileSpec {
	return readTemplateNames("files")
}

func OptionalFileNames(withDevcontainer bool) []FileSpec {
	if !withDevcontainer {
		return nil
	}
	return readTemplateNames("files/optional")
}

// readTemplateNames centraliza la lectura y filtrado de archivos .tmpl
func readTemplateNames(dir string) []FileSpec {
	files, err := templatesFS.ReadDir(dir)
	if err != nil {
		return nil
	}

	var specs []FileSpec
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".tmpl") {
			cleanName := strings.TrimSuffix(file.Name(), ".tmpl")

			targetDir := ""
			if overrideDir, exists := targetOverrides[cleanName]; exists {
				targetDir = overrideDir
			}

			specs = append(specs, FileSpec{
				Name: cleanName,
				Dir:  targetDir,
			})
		}
	}
	return specs
}

// RenderAll renderiza todas las plantillas en memoria RAM.
// Solo si el 100% de las plantillas se procesan con éxito, devuelve el mapa completo.
// Si una sola plantilla falla, aborta y no devuelve nada (semántica todo-o-nada, ADR-030).
// La clave del mapa es la ruta relativa del archivo (ej. "Dockerfile", ".devcontainer/devcontainer.json").
func RenderAll(specs []FileSpec, data project.Project) (map[string][]byte, error) {
	rendered := make(map[string][]byte, len(specs))

	for _, spec := range specs {
		renderedBytes, err := Render(spec.Name, data)
		if err != nil {
			return nil, err
		}

		key := spec.Name
		if spec.Dir != "" {
			key = filepath.Join(spec.Dir, spec.Name)
		}
		rendered[key] = renderedBytes
	}

	return rendered, nil
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
