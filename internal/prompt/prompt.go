package prompt

import (
	"bufio"
	"fmt"
	"io"
	"strings"
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

// Función string que recibe el prompting de cualquier
func String(opts StringOptions) (string, error) {

	// Iniciamos el scanner con bufio
	scanner := bufio.NewScanner(opts.Stdin)

	// Declaramos el número de reintentos dado por la inicialización del struct que se recibe.
	for i := 0; i <= opts.MaxRetries; i++ {

		// Imprimimos en la salida la pregunta
		fmt.Fprint(opts.Stdout, opts.Question+" ")

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
		fmt.Fprintln(opts.Stdout, opts.ErrorMessage)

	}

	// Cuando se supera el límite de intentos, se regresa como error.
	return "", fmt.Errorf("se superó el número máximo de reintentos (%d)", opts.MaxRetries)
}
