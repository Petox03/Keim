package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"keim/internal/generator"
	"keim/internal/godetect"
	"keim/internal/project"
	"keim/internal/templates"
	"keim/internal/ui"
	"keim/internal/validator"
)

// strategyFactory mapea nombres de estrategias (de --detect o config) a constructores.
// Es el punto exacto donde []string se convierte en []VersionStrategy (ADR-023).
var strategyFactory = map[string]func(version string) godetect.VersionStrategy{
	"host": func(_ string) godetect.VersionStrategy { return godetect.NewHostDetector() },
	"manual": func(version string) godetect.VersionStrategy {
		return godetect.NewManualDetector(version)
	},
}

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

	if err := initCmd.Parse(os.Args[2:]); err != nil {
		os.Exit(2)
	}

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

	p := project.Project{
		Name: projectName,
		Path: projectPath,
	}

	// templates inspecciona directamente embed.FS para saber qué archivos genera Keim (ADR-025).
	files := templates.FileNames()

	// Validación pre-vuelo.
	// Si la ruta no existe (ErrPathNotFound), se crea el directorio y se continúa (ADR-026).
	// Si existen archivos en conflicto o la ruta no es accesible, aborta con exit 3.
	if err := validator.Validate(p.Path, files); err != nil {
		if errors.Is(err, validator.ErrPathNotFound) {
			if err := generator.CreateProjectDir(p.Path); err != nil {
				fmt.Fprintf(os.Stderr, "keim: error: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(os.Stderr, "keim: error: %v\n", err)
			os.Exit(3)
		}
	}

	// Paso 8: Resolver la cascada de detección según la bandera --detect.
	// Error de formato en el flag = error de uso (exit 2).
	strategies, err := parseDetect(*detectFlag)
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

	// Generación de archivos desde plantillas embebidas. Error de I/O = exit 1.
	if err := generator.Generate(p, files); err != nil {
		fmt.Fprintf(os.Stderr, "keim: error: %v\n", err)
		os.Exit(1)
	}

	// Imprimir el reporte final en consola utilizando ui.PrintReport (io.Writer inyectable).
	if err := ui.PrintReport(os.Stdout, p, files); err != nil {
		fmt.Fprintf(os.Stderr, "keim: error: %v\n", err)
		os.Exit(1)
	}
}

// parseDetect convierte el string de la bandera --detect en una lista ordenable de []VersionStrategy.
// Sintaxis (ADR-004):
//   --detect host,manual=1.26   → Cascada en ese orden
//   --detect host               → Solo detección local de Go
//   --detect manual=1.26        → Manual con versión explícita
//   --detect "" (sin flag)      → Default: host -> manual (sin versión explícita)
func parseDetect(raw string) ([]godetect.VersionStrategy, error) {
	if raw == "" {
		// Default de iteración 1: host → manual (sin versión explícita).
		// En iteraciones futuras esta cascada por defecto se leerá desde config.toml.
		return []godetect.VersionStrategy{
			godetect.NewHostDetector(),
			godetect.NewManualDetector(""),
		}, nil
	}

	parts := strings.Split(raw, ",")
	strategies := make([]godetect.VersionStrategy, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("estrategia vacía en la cascada --detect")
		}

		name, version, _ := strings.Cut(part, "=")

		factory, ok := strategyFactory[name]
		if !ok {
			return nil, fmt.Errorf("estrategia '%s' no reconocida", name)
		}

		strategies = append(strategies, factory(version))
	}

	return strategies, nil
}