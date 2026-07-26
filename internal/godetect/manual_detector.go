package godetect

import (
	"fmt"
)

type ManualDetector struct {
	Version string
}

func NewManualDetector(version string) *ManualDetector {
	return &ManualDetector{
		Version: version,
	}
}

func (md *ManualDetector) Detect() (string, error) {

	if md.Version == "" {
		return "", fmt.Errorf("manual detector: la versión de Go no fue especificada (modo interactivo por stdin no soportado)")
	}

	return md.Version, nil
}
