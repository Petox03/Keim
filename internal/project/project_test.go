package project_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"keim/internal/project"
)

func TestProject_ZeroValue(t *testing.T) {

	// Validar que struct vacío (valor zero) se comporte de forma segura y que sus campos sean strings vacíos por defecto, no nil pointers.
	var p project.Project

	assert.Empty(t, p.Name)
	assert.Empty(t, p.Path)
	assert.Empty(t, p.GoVersion)
}

func TestProjectStructInitialization(t *testing.T) {
	// Validar la asignación correcta de camps para asegurar el contrato de datos.
	p := project.Project{
		Name:      "KeimApp",
		Path:      "/tmp/keim",
		GoVersion: "1.26",
	}

	assert.Equal(t, "KeimApp", p.Name)
	assert.Equal(t, "/tmp/keim", p.Path)
	assert.Equal(t, "1.26", p.GoVersion)
}
