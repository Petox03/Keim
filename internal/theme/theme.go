// Package theme centraliza la paleta de colores de Keim CLI.
// ui y prompt importan estos colores para que un cambio de paleta
// se haga en un único lugar en vez de editar cada archivo por separado.
package theme

import "github.com/charmbracelet/lipgloss"

// Paleta Keim CLI.
var (
	Primary   = lipgloss.Color("#839958")
	Secondary = lipgloss.Color("#104323")
	Accent    = lipgloss.Color("#D3968C")
	Text      = lipgloss.Color("#F7F4D5")
	Muted     = lipgloss.Color("#105666")
	Warning   = lipgloss.Color("#FFB000")
)
