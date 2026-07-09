package templates

import (
	"bytes"
	"embed"
	"strings"
	"text/template"

	"keim/internal/project"
)

//go:embed files/*.tmpl
var templatesFS embed.FS // Esto significa Embed File System que es cuando incrustas múltiples archivos o carpetas enteras.

// Collección de plantillas a nivel de paquete
var cacheTemplates *template.Template

// Ejecutar antes del main el parseo de Todos los archivos
func init() {
	// Parseo TODOS los archivos una sola vez y los dejo a nivel de caché
	// Uso template.Must para que si hay un error de sintaxis en un .tmpl,
	// el programa falle inmediatamante al arrancar en lugar de dar errores en ejecución.
	cacheTemplates = template.Must(template.ParseFS(templatesFS, "files/*"))
}

// Como el paquete template es el que maneja todo lo relacionado con las plantillas del proyecto, vi coherente
// que sea este el que tenga una función que las liste completamente, así puedo usarlo en los demás archivos
// que necesiten observar, manejar, o saber del listado de plantillas que tenemos.
func FileNames() []string {
	var names []string
	for _, tmpl := range cacheTemplates.Templates() {
		cleanName := strings.TrimSuffix(tmpl.Name(), ".tmpl")
		names = append(names, cleanName)
	}
	return names
}

func Render(name string, data project.Project) ([]byte, error){

	// 1.- Crerar colección de moldes (parseo de archivos desde templatesFS)
	// Go toma este sistema de archivos embebido y lee todo lo que esté en la carpeta
	/* templates, err := template.ParseFS(templatesFS, "files/*")
	if err != nil {
		return nil, err
	} */

	originalName := name + ".tmpl"

	// 2.- preparar la memoria con bytes.Buffer a modod e "plantilla invisible"
	// Este es el lugar intermedio donde el template puede escribir el texto generado.
	// Esto implementa io.Writter, así que es perfecto.
	var templateResult bytes.Buffer

	// 3.- Estampar los datos usando la colección que YA está en memoria
	// Acá la colección busca la plantilla con el nombre 'name' (Ojo: 'name' debe coincidir con el nombre base del archivo [ej. Dockerfile.tmpl]),
	// inyectando data y escribe el resultado en la dirección de memoria de templateResult
	err := cacheTemplates.ExecuteTemplate(&templateResult, originalName, data)
	if err != nil {
		return nil, err
	}

	// 4.- Retornamos el resultado de los bytes y como no hubo error hasta acá null en el error
	// El buffer tiene un método llamado .Bytes que da exactamente lo que necesito.
	return templateResult.Bytes(), nil
}