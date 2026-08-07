package godetect

import (
	"fmt"
	"io"
	"os"
	"regexp"

	"keim/internal/prompt"
)

type ManualDetector struct {
	Version string
	stdin   io.Reader
	stdout  io.Writer
}

func NewManualDetector(version string, stdin io.Reader, stdout io.Writer) *ManualDetector {

	if stdin == nil {
		stdin = os.Stdin
	}
	if stdout == nil {
		stdout = os.Stdout
	}

	return &ManualDetector{
		Version: version,
		stdin:   stdin,
		stdout:  stdout,
	}
}

var versionRegex = regexp.MustCompile(`^\d+\.\d+(\.\d+)?$`)

func (md *ManualDetector) Detect() (string, error) {

	if md.Version != "" {
		return md.Version, nil
	}

	res, err := prompt.String(prompt.StringOptions{
		Stdin:        md.stdin,
		Stdout:       md.stdout,
		Question:     "Ingresa la versión:",
		ErrorMessage: "Versión inválida. Intenta de nuevo.",
		MaxRetries:   2,
		Validate: func(v string) bool {
			return versionRegex.Match([]byte(v))
		},
	})

	if err != nil {
		return "", fmt.Errorf("manual detector: %w", err)
	}

	return res, nil
}
