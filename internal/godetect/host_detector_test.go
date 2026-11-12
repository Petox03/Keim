package godetect

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHostDetector_Detect(t *testing.T) {
	// Todo mock aserta el contrato del comando ejecutado: "go env GOVERSION",
	// no solo el parsing de la salida (ADR-047).
	mockGOVERSION := func(output []byte, err error) func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return func(ctx context.Context, name string, args ...string) ([]byte, error) {
			assert.Equal(t, "go", name)
			assert.Equal(t, []string{"env", "GOVERSION"}, args)
			return output, err
		}
	}

	tests := []struct {
		CaseName          string
		expectedVersion   string
		ExpectedErrSuffix string // "" significa que se espera éxito (nil)
		mockExec          func(ctx context.Context, name string, args ...string) ([]byte, error)
		TimeoutOverride   time.Duration
	}{
		{
			CaseName:        "Standard release with a line break",
			expectedVersion: "1.26.5",
			mockExec:        mockGOVERSION([]byte("go1.26.5\n"), nil),
		},
		{
			CaseName:        "Release without patch (pre-1.21 format)",
			expectedVersion: "1.20",
			mockExec:        mockGOVERSION([]byte("go1.20\n"), nil),
		},
		{
			// Regresión del bug original: un toolchain compilado con GOEXPERIMENT
			// produce "go1.26.5-X:nodwarf5" (separador "-", Go >=1.26). El parser
			// anterior aceptaba el sufijo y retornaba "1.26.5-X:nodwarf5", que se
			// propagaba a go.mod ("go 1.26.5-X:nodwarf5", sintaxis inválida) y al
			// Dockerfile ("golang:1.26.5-X:nodwarf5-alpine", imagen inexistente).
			CaseName:        "Toolchain with GOEXPERIMENT, Go >=1.26 format (reported bug)",
			expectedVersion: "1.26.5",
			mockExec:        mockGOVERSION([]byte("go1.26.5-X:nodwarf5\n"), nil),
		},
		{
			CaseName:        "Toolchain with GOEXPERIMENT, legacy Go <=1.25 format (space separator)",
			expectedVersion: "1.25.3",
			mockExec:        mockGOVERSION([]byte("go1.25.3 X:nodwarf5\n"), nil),
		},
		{
			CaseName:        "Multiple experiments in a single -X: suffix",
			expectedVersion: "1.26.5",
			mockExec:        mockGOVERSION([]byte("go1.26.5-X:nodwarf5,arenas\n"), nil),
		},
		{
			CaseName:        "Vendor/distro suffix with hyphen",
			expectedVersion: "1.21.0",
			mockExec:        mockGOVERSION([]byte("go1.21.0-1ubuntu3\n"), nil),
		},
		{
			CaseName:        "Vendor suffix with multiple hyphens",
			expectedVersion: "1.21.0",
			mockExec:        mockGOVERSION([]byte("go1.21.0-1ubuntu3-repack2\n"), nil),
		},
		{
			CaseName:        "Trailing hyphen with no suffix",
			expectedVersion: "1.26.5",
			mockExec:        mockGOVERSION([]byte("go1.26.5-\n"), nil),
		},
		{
			CaseName:        "Output with spaces, tabs and CRLF (Windows)",
			expectedVersion: "1.26.5",
			mockExec:        mockGOVERSION([]byte("  go1.26.5\t\r\n"), nil),
		},
		{
			CaseName:          "Release candidate -> cascade (product decision)",
			expectedVersion:   "",
			ExpectedErrSuffix: "salida inesperada de 'go env GOVERSION':",
			mockExec:          mockGOVERSION([]byte("go1.26rc1\n"), nil),
		},
		{
			CaseName:          "Release candidate with experiment -> cascade",
			expectedVersion:   "",
			ExpectedErrSuffix: "salida inesperada de 'go env GOVERSION':",
			mockExec:          mockGOVERSION([]byte("go1.26rc1-X:foo\n"), nil),
		},
		{
			CaseName:          "Beta -> cascade (product decision)",
			expectedVersion:   "",
			ExpectedErrSuffix: "salida inesperada de 'go env GOVERSION':",
			mockExec:          mockGOVERSION([]byte("go1.26beta2\n"), nil),
		},
		{
			// Formato devel vigente (cmd/dist post-golang/go#73372): la base
			// numérica "1.27" pasaría el regex tras cortar en "-", pero un
			// toolchain devel no es un release — el sufijo "-devel" se rechaza
			// explícitamente. Sin este chequeo, el scaffold generaría
			// "golang:1.27-alpine", imagen que no existe hasta el release.
			CaseName:          "Devel current format (go1.X-devel_hash date) -> cascade",
			expectedVersion:   "",
			ExpectedErrSuffix: "salida inesperada de 'go env GOVERSION':",
			mockExec:          mockGOVERSION([]byte("go1.27-devel_36863d6194 2025-10-17 01:07:59 -0700\n"), nil),
		},
		{
			CaseName:          "Devel intermediate format (devel go1.X-hash date) -> cascade",
			expectedVersion:   "",
			ExpectedErrSuffix: "salida inesperada de 'go env GOVERSION':",
			mockExec:          mockGOVERSION([]byte("devel go1.26-abc1234 Tue Oct 20 16:43:37 2020 -0400\n"), nil),
		},
		{
			CaseName:          "Devel old format (devel +hash date) -> cascade",
			expectedVersion:   "",
			ExpectedErrSuffix: "salida inesperada de 'go env GOVERSION':",
			mockExec:          mockGOVERSION([]byte("devel +afb5fca Sun Aug 07 00:00:00 2020 +0000\n"), nil),
		},
		{
			CaseName:          "Devel with chained experiment -> cascade",
			expectedVersion:   "",
			ExpectedErrSuffix: "salida inesperada de 'go env GOVERSION':",
			mockExec:          mockGOVERSION([]byte("go1.27-devel_abc1234-X:foo\n"), nil),
		},
		{
			// BoringCrypto ("go1.21.0b7") es un release real, pero su sufijo "b7"
			// sin separador no pasa el regex → cascada. Limitación aceptada y
			// documentada en ADR-047.
			CaseName:          "BoringCrypto (bN suffix) -> cascade (documented limitation)",
			expectedVersion:   "",
			ExpectedErrSuffix: "salida inesperada de 'go env GOVERSION':",
			mockExec:          mockGOVERSION([]byte("go1.21.0b7\n"), nil),
		},
		{
			CaseName:          "Empty output",
			expectedVersion:   "",
			ExpectedErrSuffix: "salida inesperada de 'go env GOVERSION':",
			mockExec:          mockGOVERSION([]byte(""), nil),
		},
		{
			CaseName:          "Line break only",
			expectedVersion:   "",
			ExpectedErrSuffix: "salida inesperada de 'go env GOVERSION':",
			mockExec:          mockGOVERSION([]byte("\n"), nil),
		},
		{
			CaseName:          "Bare 'go' token without version",
			expectedVersion:   "",
			ExpectedErrSuffix: "salida inesperada de 'go env GOVERSION':",
			mockExec:          mockGOVERSION([]byte("go\n"), nil),
		},
		{
			CaseName:          "Garbage without go prefix",
			expectedVersion:   "",
			ExpectedErrSuffix: "salida inesperada de 'go env GOVERSION':",
			mockExec:          mockGOVERSION([]byte("Broken command"), nil),
		},
		{
			CaseName:          "Prefix-like string not rescued (godevel)",
			expectedVersion:   "",
			ExpectedErrSuffix: "salida inesperada de 'go env GOVERSION':",
			mockExec:          mockGOVERSION([]byte("godevel\n"), nil),
		},
		{
			CaseName:          "Run time error: the 'go' command doesn't exist",
			expectedVersion:   "",
			ExpectedErrSuffix: "error al ejecutar el comando",
			mockExec:          mockGOVERSION(nil, errors.New("executable file not found in $PATH")),
		},
		{
			CaseName:          "Timeout / Context Exceeded",
			expectedVersion:   "",
			ExpectedErrSuffix: "error al ejecutar el comando",
			mockExec: func(ctx context.Context, name string, args ...string) ([]byte, error) {
				assert.Equal(t, "go", name)
				assert.Equal(t, []string{"env", "GOVERSION"}, args)
				select {
				case <-time.After(50 * time.Millisecond):
					return []byte("go1.26.5\n"), nil
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			},
			TimeoutOverride: 10 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.CaseName, func(t *testing.T) {
			// 1. Inyectar dependencias
			hd := &HostDetector{
				execFn:  tt.mockExec,
				timeout: 500 * time.Millisecond,
			}

			if tt.TimeoutOverride > 0 {
				hd.timeout = tt.TimeoutOverride
			}

			// 2. Ejecución
			version, err := hd.Detect()

			// 3. Evaluamos los resultados
			if tt.ExpectedErrSuffix == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedVersion, version)
			} else {
				assert.ErrorContains(t, err, tt.ExpectedErrSuffix)
			}
		})
	}
}
