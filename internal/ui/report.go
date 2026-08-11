package ui

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"keim/internal/project"
	"keim/internal/theme"

	"github.com/charmbracelet/lipgloss"
)

func PrintReport(w io.Writer, p project.Project, files []string) error {
	if w == nil {
		return fmt.Errorf("el io.Writer no puede ser nil")
	}

	// Estilos basados en la paleta centralizada de Keim CLI (internal/theme).
	styleTitle := lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)
	styleHeader := lipgloss.NewStyle().Foreground(theme.Text).Bold(true)
	styleLabel := lipgloss.NewStyle().Foreground(theme.Muted)
	styleValue := lipgloss.NewStyle().Foreground(theme.Text).Bold(true)
	styleBullet := lipgloss.NewStyle().Foreground(theme.Accent)
	styleCode := lipgloss.NewStyle().Foreground(theme.Primary)
	styleBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Secondary).
		Padding(1, 2)

	var sb strings.Builder

	// Encabezado principal
	fmt.Fprintf(&sb, "%s\n\n", styleTitle.Render(fmt.Sprintf("¡Proyecto '%s' creado con éxito!", p.Name)))

	// Detalles del proyecto
	fmt.Fprintf(&sb, "%s\n", styleHeader.Render("Detalles del proyecto:"))
	fmt.Fprintf(&sb, "  %s %s %s\n", styleBullet.Render("•"), styleLabel.Render("Ruta:           "), styleValue.Render(p.Path))
	fmt.Fprintf(&sb, "  %s %s %s\n", styleBullet.Render("•"), styleLabel.Render("Versión de Go:  "), styleValue.Render(p.GoVersion))
	fmt.Fprintf(&sb, "  %s %s %s\n\n", styleBullet.Render("•"), styleLabel.Render("Devcontainer:   "), styleValue.Render(fmt.Sprintf("%v", p.WithDevcontainer)))

	// Archivos generados
	fmt.Fprintf(&sb, "%s\n", styleHeader.Render("Archivos generados:"))
	for _, file := range files {
		fmt.Fprintf(&sb, "  %s %s\n", styleBullet.Render("•"), styleLabel.Render(file))
	}
	sb.WriteString("\n")

	// Siguientes pasos
	fmt.Fprintf(&sb, "%s\n", styleHeader.Render("Siguientes pasos:"))
	cleanPath := filepath.Clean(p.Path)

	if cleanPath == "." || cleanPath == "" {
		fmt.Fprintf(&sb, "  %s %s\n", styleBullet.Render("1."), styleCode.Render("docker compose up -d"))
		fmt.Fprintf(&sb, "  %s %s\n", styleBullet.Render("2."), styleCode.Render("docker compose exec app go run ."))
	} else {
		fmt.Fprintf(&sb, "  %s %s\n", styleBullet.Render("1."), styleCode.Render("cd "+cleanPath))
		fmt.Fprintf(&sb, "  %s %s\n", styleBullet.Render("2."), styleCode.Render("docker compose up -d"))
		fmt.Fprintf(&sb, "  %s %s\n", styleBullet.Render("3."), styleCode.Render("docker compose exec app go run ."))
	}

	_, err := fmt.Fprintln(w, styleBox.Render(sb.String()))
	return err
}
