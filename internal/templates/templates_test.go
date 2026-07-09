package templates_test

import(
	"testing"

	"github.com/stretchr/testify/assert"
	"keim/internal/project"
	"keim/internal/templates"
)

func TestFileNames(t *testing.T) {
	// Verificar que al menos descubra dinámicamente los archivos MVP
	files := templates.FileNames()

	assert.Contains(t, files, "go.mod")
	assert.Contains(t, files, "main.go")
	assert.Contains(t, files, "Dockerfile")
}

func TestRenderSuccess(t *testing.T) {
	p := project.Project{
		Name:		"testapp",
		GoVersion:	"1.26",
	}

	// Renderizar go.mod
	bytesResult, err := templates.Render("go.mod", p)

	assert.NoError(t, err)
	assert.NotEmpty(t, bytesResult)

	// Verificar que las variables del struct se hayan estampado
	strContent := string(bytesResult)
	assert.Contains(t, strContent, "module testapp")
	assert.Contains(t, strContent, "go 1.26")
}

func TestRenderNotFound(t *testing.T) {
	p := project.Project{ Name: "test" }

	// Pedir un archivo que no existe debe retornar un error
	_, err := templates.Render("test", p)

	assert.Error(t, err)
}