package prompt

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	isNotEmpty := func(s string) bool { return s != "" }

	tests := []struct {
		CaseName       string
		SimulatedInput string // lo que el usuario "escribe", línea por línea
		MaxRetries     int
		ExpectedResult string
		ExpectError    bool
	}{
		{
			CaseName:       "Success on first try",
			SimulatedInput: "1.26\n",
			MaxRetries:     2,
			ExpectedResult: "1.26",
			ExpectError:    false,
		},
		{
			CaseName:       "Fails once, then succeeds on retry",
			SimulatedInput: "\n1.26\n", // primera línea vacía (inválida), segunda válida
			MaxRetries:     2,
			ExpectedResult: "1.26",
			ExpectError:    false,
		},
		{
			CaseName:       "Exhausts all retries without valid input",
			SimulatedInput: "\n\n\n", // tres líneas vacías, todas inválidas
			MaxRetries:     2,        // 1 intento inicial + 2 reintentos = 3 intentos totales
			ExpectedResult: "",
			ExpectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.CaseName, func(t *testing.T) {
			stdin := strings.NewReader(tt.SimulatedInput)
			var stdout bytes.Buffer

			result, err := String(StringOptions{
				Stdin:        stdin,
				Stdout:       &stdout,
				Question:     "Enter version: ",
				ErrorMessage: "Invalid input",
				MaxRetries:   tt.MaxRetries,
				Validate:     isNotEmpty,
			})

			if tt.ExpectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.ExpectedResult, result)
		})
	}
}

func TestString_StdoutContent(t *testing.T) {
	t.Run("Question appears in stdout", func(t *testing.T) {
		stdin := strings.NewReader("1.26\n")
		var stdout bytes.Buffer

		_, err := String(StringOptions{
			Stdin:        stdin,
			Stdout:       &stdout,
			Question:     "Ingresa la versión:",
			ErrorMessage: "Versión inválida",
			MaxRetries:   2,
			Validate:     func(s string) bool { return s != "" },
		})

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "Ingresa la versión:")
	})

	t.Run("Error message appears in stdout on invalid input", func(t *testing.T) {
		stdin := strings.NewReader("\n1.26\n")
		var stdout bytes.Buffer

		_, err := String(StringOptions{
			Stdin:        stdin,
			Stdout:       &stdout,
			Question:     "Ingresa la versión:",
			ErrorMessage: "Versión inválida",
			MaxRetries:   2,
			Validate:     func(s string) bool { return s != "" },
		})

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "Versión inválida")
	})
}

func TestString_EOF(t *testing.T) {
	isNotEmpty := func(s string) bool { return s != "" }

	t.Run("Empty stdin aborts immediately on EOF", func(t *testing.T) {
		stdin := strings.NewReader("")
		var stdout bytes.Buffer

		result, err := String(StringOptions{
			Stdin:        stdin,
			Stdout:       &stdout,
			Question:     "Enter version: ",
			ErrorMessage: "Invalid input",
			MaxRetries:   2,
			Validate:     isNotEmpty,
		})

		assert.Error(t, err)
		assert.ErrorIs(t, err, io.EOF)
		assert.Empty(t, result)
		assert.Equal(t, 1, strings.Count(stdout.String(), "Enter version:"),
			"la pregunta debe imprimirse una sola vez, no una por reintento")
		assert.NotContains(t, stdout.String(), "Invalid input")
	})

	t.Run("Aborts on EOF after an invalid line", func(t *testing.T) {
		stdin := strings.NewReader("abc\n")
		var stdout bytes.Buffer

		_, err := String(StringOptions{
			Stdin:        stdin,
			Stdout:       &stdout,
			Question:     "Enter version: ",
			ErrorMessage: "Invalid input",
			MaxRetries:   2,
			Validate:     func(s string) bool { return s == "1.26" },
		})

		assert.Error(t, err)
		assert.ErrorIs(t, err, io.EOF)
		assert.Equal(t, 2, strings.Count(stdout.String(), "Enter version:"),
			"el error debe llegar en el segundo intento, no tras agotar reintentos")
		assert.Equal(t, 1, strings.Count(stdout.String(), "Invalid input"))
	})

	t.Run("Valid input without trailing newline succeeds", func(t *testing.T) {
		stdin := strings.NewReader("1.26")
		var stdout bytes.Buffer

		result, err := String(StringOptions{
			Stdin:        stdin,
			Stdout:       &stdout,
			Question:     "Enter version: ",
			ErrorMessage: "Invalid input",
			MaxRetries:   2,
			Validate:     isNotEmpty,
		})

		assert.NoError(t, err)
		assert.Equal(t, "1.26", result)
	})
}

func TestConfirm(t *testing.T) {
	tests := []struct {
		CaseName       string
		SimulatedInput string
		MaxRetries     int
		ExpectedResult bool
		ExpectError    bool
	}{
		{
			CaseName:       "Yes on first try",
			SimulatedInput: "y\n",
			MaxRetries:     2,
			ExpectedResult: true,
			ExpectError:    false,
		},
		{
			CaseName:       "Yes full word",
			SimulatedInput: "yes\n",
			MaxRetries:     2,
			ExpectedResult: true,
			ExpectError:    false,
		},
		{
			CaseName:       "No on first try",
			SimulatedInput: "n\n",
			MaxRetries:     2,
			ExpectedResult: false,
			ExpectError:    false,
		},
		{
			CaseName:       "No full word",
			SimulatedInput: "no\n",
			MaxRetries:     2,
			ExpectedResult: false,
			ExpectError:    false,
		},
		{
			CaseName:       "Case insensitive YES",
			SimulatedInput: "YES\n",
			MaxRetries:     2,
			ExpectedResult: true,
			ExpectError:    false,
		},
		{
			CaseName:       "Invalid then yes on retry",
			SimulatedInput: "maybe\ny\n",
			MaxRetries:     2,
			ExpectedResult: true,
			ExpectError:    false,
		},
		{
			CaseName:       "Exhausts all retries",
			SimulatedInput: "maybe\nmaybe\nmaybe\n",
			MaxRetries:     2,
			ExpectedResult: false,
			ExpectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.CaseName, func(t *testing.T) {
			stdin := strings.NewReader(tt.SimulatedInput)
			var stdout bytes.Buffer

			result, err := Confirm(ConfirmOptions{
				Stdin:        stdin,
				Stdout:       &stdout,
				Question:     "¿Usar devcontainer? [y/n]",
				ErrorMessage: "Respuesta inválida. Use 'y' o 'n'.",
				MaxRetries:   tt.MaxRetries,
			})

			if tt.ExpectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.ExpectedResult, result)
		})
	}
}
