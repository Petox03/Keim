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
	return &HostDetector{
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

	// El output suele traer un salto de línea al final, Limpiar y separar
	rawVersion := strings.TrimSpace(string(output))
	words := strings.Fields(rawVersion)

	for _, word := range words {
		if strings.HasPrefix(word, "go") && len(word) > 2 && word[2] >= '0' && word[2] <= '9' {
			return word[2:], nil
		}
	}

	return "", fmt.Errorf("no se encontró una versión válida de Go en la salida: %s", rawVersion)
}
