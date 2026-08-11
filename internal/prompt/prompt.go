package prompt

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"keim/internal/theme"

	"github.com/charmbracelet/lipgloss"
)

// Struct usado a modo de DTO porque son muchos parámetros.
type StringOptions struct {
	Stdin        io.Reader
	Stdout       io.Writer
	Question     string
	ErrorMessage string
	MaxRetries   int
	Validate     func(string) bool
}

// ConfirmOptions configura una pregunta de sí/no.
type ConfirmOptions struct {
	Stdin        io.Reader
	Stdout       io.Writer
	Question     string
	ErrorMessage string
	MaxRetries   int
}

// isYesNo valida que la entrada sea y/yes/n/no (case-insensitive).
func isYesNo(input string) bool {
	switch strings.ToLower(input) {
	case "y", "yes", "n", "no":
		return true
	default:
		return false
	}
}

// Confirm hace una pregunta de sí/no y devuelve un bool.
// Reusa la maquinaria de String con un validador interno para y/yes/n/no.
func Confirm(opts ConfirmOptions) (bool, error) {
	result, err := String(StringOptions{
		Stdin:        opts.Stdin,
		Stdout:       opts.Stdout,
		Question:     opts.Question,
		ErrorMessage: opts.ErrorMessage,
		MaxRetries:   opts.MaxRetries,
		Validate:     isYesNo,
	})
	if err != nil {
		return false, err
	}

	return strings.HasPrefix(strings.ToLower(result), "y"), nil
}

// Función string que recibe el prompting de cualquier
func String(opts StringOptions) (string, error) {

	stylePrompt := lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)
	styleError := lipgloss.NewStyle().Foreground(theme.Warning)

	// Iniciamos el scanner con bufio
	scanner := bufio.NewScanner(opts.Stdin)

	// Declaramos el número de reintentos dado por la inicialización del struct que se recibe.
	for i := 0; i <= opts.MaxRetries; i++ {

		// Imprimimos en la salida la pregunta con estilo
		fmt.Fprint(opts.Stdout, stylePrompt.Render("? ")+opts.Question+" ")

		// Scan() Devuelve un bool. Si se hace el escanoe pasa, sino se lo salta
		if scanner.Scan() {

			// Quitamos los espacios de el texto guardado en memoria
			input := strings.TrimSpace(scanner.Text())

			// Si no hay que validar o la validación pasa, regresamos el texto sin errores
			if opts.Validate == nil || opts.Validate(input) {
				return input, nil
			}
		}

		// Si el escaner dio error lo regresamos.
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("error al leer la entrada: %w", err)
		}

		// Si dio un formato inválido, le regresamos en la tubería de la salida dada el mensaje de error
		fmt.Fprintln(opts.Stdout, styleError.Render(opts.ErrorMessage))

	}

	// Cuando se supera el límite de intentos, se regresa como error.
	return "", fmt.Errorf("se superó el número máximo de reintentos (%d)", opts.MaxRetries)
}
