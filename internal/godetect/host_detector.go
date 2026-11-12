package godetect

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type HostDetector struct {
	execFn  func(ctx context.Context, name string, args ...string) ([]byte, error)
	timeout time.Duration
}

func NewHostDetector() *HostDetector {
	return &HostDetector{
		execFn: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return exec.CommandContext(ctx, name, args...).Output()
		},
		timeout: 500 * time.Millisecond,
	}
}

func (hd *HostDetector) Detect() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), hd.timeout)
	defer cancel()

	// "go env GOVERSION" devuelve runtime.Version() en formato machine-readable (ADR-047)
	output, err := hd.execFn(ctx, "go", "env", "GOVERSION")
	if err != nil {
		return "", fmt.Errorf("error al ejecutar el comando: %w", err)
	}

	raw := strings.TrimSpace(string(output))
	fields := strings.Fields(raw)

	var version string
	if len(fields) > 0 {
		version = strings.TrimPrefix(fields[0], "go")
	}

	// Cortar en el primer guion recupera la base de sufijos de release
	// ("-X:nodwarf5" desde Go 1.26, vendor como "-1ubuntu3"). El sufijo
	// "-devel" es la excepción: un toolchain devel no es un release aunque
	// su base numérica sea válida, y cae a la cascada (ADR-047).
	base, suffix, _ := strings.Cut(version, "-")
	if strings.HasPrefix(suffix, "devel") || !versionRegex.MatchString(base) {
		return "", fmt.Errorf("salida inesperada de 'go env GOVERSION': %s", raw)
	}
	return base, nil
}
