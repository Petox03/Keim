package godetect

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostDetector_Detect(t *testing.T) {
	tests := []struct {
		CaseName          string
		expectedVersion   string
		ExpectedErrSuffix string // "" significa que se espera éxito (nil)
		mockExec          func(name string, args ...string) ([]byte, error)
	}{
		{
			CaseName:          "Correct output with a line break",
			expectedVersion:   "1.26.2",
			ExpectedErrSuffix: "",
			mockExec: func(name string, args ...string) ([]byte, error) {
				return []byte("go version go1.26.2 linux/amd64\n"), nil
			},
		},
		{
			CaseName:          "Run time error: The 'go' command doesn't exist",
			expectedVersion:   "",
			ExpectedErrSuffix: "error al ejecutar el comando",
			mockExec: func(name string, args ...string) ([]byte, error) {
				return nil, errors.New("executable file not found in $PATH")
			},
		},
		{
			CaseName:          "Unexpected format: fewer than 3 words",
			expectedVersion:   "",
			ExpectedErrSuffix: "no se encontró una versión válida de Go en la salida:",
			mockExec: func(name string, args ...string) ([]byte, error) {
				return []byte("Broken command"), nil
			},
		},
		{
			CaseName:          "Custom or modified output (devel / extra words)",
			expectedVersion:   "1.23.0",
			ExpectedErrSuffix: "",
			mockExec: func(name string, args ...string) ([]byte, error) {
				// Simula salidas con prefijos o detalles extra antes/después de la versión
				return []byte("go version devel go1.23.0 custom-build linux/amd64\n"), nil
			},
		},
		{
			CaseName:          "Malformed go prefix (no digits after 'go')",
			expectedVersion:   "",
			ExpectedErrSuffix: "no se encontró una versión válida de Go en la salida:",
			mockExec: func(name string, args ...string) ([]byte, error) {
				return []byte("go version godevel linux/amd64\n"), nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.CaseName, func(t *testing.T) {
			// 1. Inyectar dependencias
			hd := &HostDetector{
				execFn: tt.mockExec,
			}

			// 2. Ejecución
			version, err := hd.Detect()

			// 3. Evaluamos los resultados
			if tt.ExpectedErrSuffix == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedVersion, version)
			} else {
				//assert.Error(t, err)
				assert.ErrorContains(t, err, tt.ExpectedErrSuffix)
			}
		})
	}
}
