package godetect

import (
	"fmt"
	"strings"
)

// strategyFactory mapea nombres de estrategias (de --detect o config) a constructores.
// Es el punto exacto donde []string se convierte en []VersionStrategy (ADR-023).
var strategyFactory = map[string]func(version string) VersionStrategy{
	"host": func(_ string) VersionStrategy { return NewHostDetector() },
	"internet": func(_ string) VersionStrategy {
		return NewInternetDetector(nil) // nil → *http.Client{} por defecto (ver NewInternetDetector)
	},
	"manual": func(version string) VersionStrategy {
		return NewManualDetector(version, nil, nil) // nil, nil → os.Stdin, os.Stdout por defecto (ver NewManualDetector)
	},
}

// parseDetect convierte el string de la bandera --detect en una lista ordenable de []VersionStrategy.
// Sintaxis (ADR-004):
//
//	--detect host,manual=1.26   → Cascada en ese orden
//	--detect host               → Solo detección local de Go
//	--detect manual=1.26        → Manual con versión explícita
//	--detect "" (sin flag)      → Default: host → internet → manual
func ParseDetect(raw string) ([]VersionStrategy, error) {
	if raw == "" {
		// Default de la cascada
		return ParseDetect("host,internet,manual")
	}

	parts := strings.Split(raw, ",")
	strategies := make([]VersionStrategy, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("estrategia vacía en la cascada --detect")
		}

		name, version, _ := strings.Cut(part, "=")

		factory, ok := strategyFactory[name]
		if !ok {
			return nil, fmt.Errorf("estrategia '%s' no reconocida", name)
		}

		strategies = append(strategies, factory(version))
	}

	return strategies, nil
}
