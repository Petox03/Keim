package godetect

// VersionStrategy define el contrato que toda estrategia de detección de versión de Go debe cumplir.
type VersionStrategy interface {
	Detect() (string, error)
}
