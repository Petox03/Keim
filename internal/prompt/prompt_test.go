package prompt

import (
	"bytes"
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
