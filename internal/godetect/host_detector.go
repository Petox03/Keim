package godetect

import (
	"fmt"
	"os/exec"
	"strings"
)

type HostDetector struct {
    execFn func(name string, args ...string) ([]byte, error)
}

func NewHostDetector() *HostDetector {
    return &HostDetector {
        execFn: func(name string, args ...string) ([]byte, error) {
			return exec.Command(name, args...).Output()
		},
    }
}

func (hd *HostDetector) Detect() (string, error) {
	// Ejecutamos "go version" a través de nuestro wrapper execFn
	output, err := hd.execFn("go", "version")
	if err != nil {
		return "", fmt.Errorf("error al ejecutar el comando: %w", err)
	}

    // TODO: parsing frágil por índice, buscar la palabra que empieza con 'go' es mejor.

	// El output suele traer un salto de línea al final, Limpiar y separar
	rawVersion := strings.TrimSpace(string(output))
    words := strings.Fields(rawVersion)

    if len(words) < 3 {
        return "", fmt.Errorf("la salida del comando no tiene el formato esperado: %s", rawVersion)
    }

    version := strings.TrimPrefix(words[2], "go")

	return version, nil
}