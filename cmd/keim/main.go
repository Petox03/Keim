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
	// Paso 1-2: verificar que el comando sea "init".
	// Keim en esta iteración solo entiende "init". Cualquier otra cosa es error de uso (exit 2).
	if len(os.Args) < 2 || os.Args[1] != "init" {
		fmt.Fprintln(os.Stderr, "uso: keim init [--detect <cascada>] [nombre]")
		os.Exit(2)
	}

	// Paso 3: sacar "init" de os.Args para que flag.Parse() no se confunda con él.
	// Nota (ADR-027): esto NO permite flags en cualquier posición. flag.Parse() se
	// detiene en el primer argumento posicional. El usuario debe poner --detect
	// ANTES del nombre: "keim init --detect host clippy".
	os.Args = append(os.Args[:1], os.Args[2:]...)

	// Paso 4-5: declarar y parsear flags.
	detectFlag := flag.String("detect", "", "cascada de detección (ej: host,manual=1.26)")
	flag.Parse()

	// Paso 6: obtener el nombre del proyecto si lo hay (sobrante tras parsear flags).
	// En esta iteración Keim acepta máximo 1 argumento posicional (el nombre del
	// proyecto). Si hay más, es error de uso: el usuario probablemente puso --detect
	// después del nombre (flag.Parse() se detiene ahí y lo deja como posicional) o
	// pasó varios nombres. En cualquier caso, mejor un mensaje claro que un flag
	// silenciosamente ignorado que termina en "detección fallida" engañoso.
	args := flag.Args()
	if len(args) > 1 {
		fmt.Fprintln(os.Stderr, "keim: error: demasiados argumentos posicionales. Uso: keim init [--detect <cascada>] [nombre]")
		os.Exit(2)
	}
	var projectName string
	if len(args) == 1 {
		projectName = args[0]
	}

	// Paso 7: resolver Name y Path según haya o no nombre.
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

	// Fuente única de verdad: templates sabe qué archivos genera Keim (ADR-025).
	files := templates.FileNames()

	// Validación pre-vuelo.
	// Si la ruta no existe (ErrPathNotFound), se crea y se continúa (ADR-026).
	// Si hay conflictos o la ruta no es accesible, aborta con exit 3.
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

	// Paso 8: resolver la cascada de detección según --detect.
	// Error de parsing del flag = error de uso (exit 2).
	strategies, err := parseDetect(*detectFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "keim: error: %v\n", err)
		os.Exit(2)
	}

	// Detección de versión de Go. Si todas las estrategias fallan, exit 4.
	version, err := godetect.Detect(strategies)
	if err != nil {
		fmt.Fprintf(os.Stderr, "keim: error: detección fallida: %v\n", err)
		os.Exit(4)
	}

	p.GoVersion = version

	// Generación de archivos. Error de IO = exit 1.
	if err := generator.Generate(p, files); err != nil {
		fmt.Fprintf(os.Stderr, "keim: error: %v\n", err)
		os.Exit(1)
	}

	// Reporte final por ui.PrintReport (commit 6, io.Writer inyectable).
	if err := ui.PrintReport(os.Stdout, p, files); err != nil {
		fmt.Fprintf(os.Stderr, "keim: error: %v\n", err)
		os.Exit(1)
	}
}

// parseDetect convierte el valor de --detect en []VersionStrategy.
// Sintaxis (ADR-004):
//   --detect host,manual=1.26   → cascada en ese orden
//   --detect host               → solo host
//   --detect manual=1.26        → manual con versión explícita
//   --detect "" (sin flag)      → default: host,manual (sin versión explícita)
func parseDetect(raw string) ([]godetect.VersionStrategy, error) {
	if raw == "" {
		// Default de iteración 1: host → manual (sin versión explícita).
		// En iteración 2 esto se lee de config.toml.
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
