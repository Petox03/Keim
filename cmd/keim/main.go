package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"keim/internal/generator"
	"keim/internal/godetect"
	"keim/internal/project"
	"keim/internal/prompt"
	"keim/internal/templates"
	"keim/internal/ui"
	"keim/internal/validator"
)

func main() {
	// Paso 1-2: Verificar que el primer argumento posicional sea el subcomando "init".
	// Keim en esta iteración solo entiende "init". Cualquier otra opción o ausencia
	// de subcomando es un error de uso (exit 2).
	if len(os.Args) < 2 || os.Args[1] != "init" {
		fmt.Fprintln(os.Stderr, "uso: keim init [--detect <cascada>] [nombre]")
		os.Exit(2)
	}

	// Paso 3: Crear un FlagSet aislado para el subcomando "init".
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)

	// Paso 4-5: Registrar y parsear banderas exclusivas de "init".
	// Nota (ADR-027): flag.FlagSet detiene el parseo al encontrar el primer argumento posicional.
	// Esto exige al usuario colocar las banderas ANTES del nombre del proyecto:
	// "keim init --detect host clippy". Si lo invierte, --detect se tratará como posicional sobrante.
	detectFlag := initCmd.String("detect", "", "cascada de detección (ej: host,manual=1.26)")
	devcontainerFlag := initCmd.Bool("devcontainer", false, "generar configuración de devcontainer para VS Code/IDE compatibles")

	if err := initCmd.Parse(os.Args[2:]); err != nil {
		os.Exit(2)
	}

	// Detectar si --devcontainer fue pasado explícitamente (tri-state).
	// flag.Visit sólo recorre los flags que el usuario escribió en la línea de comandos.
	// Si no lo pasó, necesitamos decidir: prompt interactivo (si hay TTY) o default false (si no).
	devcontainerSet := false
	initCmd.Visit(func(f *flag.Flag) {
		if f.Name == "devcontainer" {
			devcontainerSet = true
		}
	})

	// Paso 6: Obtener el nombre del proyecto desde los argumentos posicionales restantes.
	// Keim acepta máximo 1 argumento posicional (el nombre). Si hay más, es error de uso:
	// el usuario probablemente puso banderas después del nombre (haciendo que el parsing se detenga)
	// o pasó múltiples nombres. Abortamos para evitar conductas engañosas.
	args := initCmd.Args()
	if len(args) > 1 {
		fmt.Fprintln(os.Stderr, "keim: error: demasiados argumentos posicionales. Uso: keim init [--detect <cascada>] [nombre]")
		os.Exit(2)
	}
	var projectName string
	if len(args) == 1 {
		projectName = args[0]
	}

	// Paso 7: Resolver Name y Path según se haya proporcionado o no un nombre.
	projectPath := "."
	if projectName == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "keim: error: no se pudo determinar el directorio actual: %v\n", err)
			os.Exit(1)
		}
		projectName = filepath.Base(cwd)
	} else {
		projectPath = filepath.Join(".", projectName)
	}

	// Resolver WithDevcontainer: tri-state (explícito true/false, o prompt interactivo).
	var withDevcontainer bool
	if devcontainerSet {
		withDevcontainer = *devcontainerFlag
	} else {
		withDevcontainer = resolveDevcontainerInteractive()
	}

	p := project.Project{
		Name:             projectName,
		Path:             projectPath,
		WithDevcontainer: withDevcontainer,
	}

	// templates inspecciona directamente embed.FS para saber qué archivos genera Keim (ADR-025).
	specs := templates.AllFileNames(p.WithDevcontainer)
	forbiddenFiles := templates.GetForbiddenFiles(p.WithDevcontainer)

	// Validación pre-vuelo.
	// Si la ruta no existe (ErrPathNotFound), se crea el directorio y se continúa (ADR-026).
	// Si existen archivos en conflicto o la ruta no es accesible, aborta con exit 3.
	needDir := false
	if err := validator.Validate(p.Path, forbiddenFiles); err != nil {
		if errors.Is(err, validator.ErrPathNotFound) {
			needDir = true
		} else {
			fmt.Fprintf(os.Stderr, "keim: error: %v\n", err)
			os.Exit(3)
		}
	}

	// Paso 8: Resolver la cascada de detección según la bandera --detect.
	// Error de formato en el flag = error de uso (exit 2).
	strategies, err := godetect.ParseDetect(*detectFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "keim: error: %v\n", err)
		os.Exit(2)
	}

	// Detección de la versión de Go. Si ninguna estrategia responde con éxito, aborta con exit 4.
	version, err := godetect.Detect(strategies)
	if err != nil {
		fmt.Fprintf(os.Stderr, "keim: error: detección fallida: %v\n", err)
		os.Exit(4)
	}

	p.GoVersion = version

	// Renderizar todas las plantillas en memoria (todo-o-nada, ADR-030).
	// Si una sola plantilla falla, abortamos sin tocar el disco.
	rendered, err := templates.RenderAll(specs, p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "keim: error: %v\n", err)
		os.Exit(1)
	}

	// Generación de archivos desde plantillas embebidas. Error de I/O = exit 1.
	if needDir {
		if err := generator.CreateProjectDir(p.Path); err != nil {
			fmt.Fprintf(os.Stderr, "keim: error: %v\n", err)
			os.Exit(1)
		}
	}
	if err := generator.WriteFiles(p, rendered); err != nil {
		fmt.Fprintf(os.Stderr, "keim: error: %v\n", err)
		os.Exit(1)
	}

	// Calcular las rutas de display para el reporte (wiring: FileSpec → string).
	displayPaths := make([]string, len(specs))
	for i, spec := range specs {
		if spec.Dir != "" {
			displayPaths[i] = filepath.Join(spec.Dir, spec.Name)
		} else {
			displayPaths[i] = spec.Name
		}
	}

	// Imprimir el reporte final en consola utilizando ui.PrintReport (io.Writer inyectable).
	if err := ui.PrintReport(os.Stdout, p, displayPaths); err != nil {
		fmt.Fprintf(os.Stderr, "keim: error: %v\n", err)
		os.Exit(1)
	}
}

// resolveDevcontainerInteractive decide si generar devcontainer cuando el usuario
// no pasó --devcontainer explícitamente. Si stdin es un TTY, pregunta interactivamente.
// Si no (CI pipeline, pipe, redirección), default false silencioso.
func resolveDevcontainerInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	// ModeCharDevice indica que stdin es una terminal interactiva.
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		return false
	}

	result, err := prompt.Confirm(prompt.ConfirmOptions{
		Stdin:        os.Stdin,
		Stdout:       os.Stdout,
		Question:     "¿Generar configuración de devcontainer? [y/n]",
		ErrorMessage: "Respuesta inválida. Use 'y' o 'n'.",
		MaxRetries:   2,
	})
	if err != nil {
		return false
	}

	return result
}
