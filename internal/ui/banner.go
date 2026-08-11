package ui

import (
	"fmt"
	"io"
	"strings"

	"keim/internal/theme"

	"github.com/charmbracelet/lipgloss"
)

// PrintBanner renderiza el banner ASCII de Keim con lipgloss.
// La versión se inyecta como parámetro para que main.go la resuelva (git tag, hardcoded, etc.).
func PrintBanner(w io.Writer, version string) error {
	if w == nil {
		return fmt.Errorf("el io.Writer no puede ser nil")
	}

	// Normalizar prefijo "v": git describe devuelve "v0.2.1", hardcodeo "0.2.1", fallback "dev".
	// "dev" no lleva prefijo. Si ya tiene "v", no se duplica.
	if version != "dev" && !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	asciiArt := `
 __  __    ______    __    __    __
/\ \/ /   /\  ___\  /\ \  /\ "-./  \
\ \  _"-. \ \  __\  \ \ \ \ \ \-./\ \
 \ \_\ \_\ \ \_____\ \ \_\ \ \_\ \ \_\
  \/_/\/_/  \/_____/  \/_/  \/_/  \/_/
`

	logoStyle := lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true)

	titleStyle := lipgloss.NewStyle().
		Foreground(theme.Text).
		Bold(true).
		SetString("KEIM CLI")

	versionStyle := lipgloss.NewStyle().
		Foreground(theme.Secondary).
		Italic(true).
		SetString(version)

	taglineStyle := lipgloss.NewStyle().
		Foreground(theme.Muted).
		SetString("Scaffolder de proyectos Go dockerizados.")

	authorStyle := lipgloss.NewStyle().
		Foreground(theme.Accent).
		SetString("by Alberto Sosa")

	headerLine := lipgloss.JoinHorizontal(
		lipgloss.Bottom,
		titleStyle.Render(),
		" ",
		versionStyle.Render(),
		"  ",
		authorStyle.Render(),
	)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		logoStyle.Render(asciiArt),
		"",
		headerLine,
		taglineStyle.Render(),
	)

	bannerBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(theme.Secondary).
		PaddingLeft(3).
		PaddingTop(1).
		PaddingBottom(1).
		Render(content)

	_, err := fmt.Fprintln(w, "\n"+bannerBox+"\n")
	return err
}
