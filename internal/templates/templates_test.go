package templates_test

import (
	"path/filepath"
	"strings"
	"testing"

	"keim/internal/project"
	"keim/internal/templates"

	"github.com/stretchr/testify/assert"
)

func TestFileNames(t *testing.T) {
	files := templates.FileNames()

	var names []string
	for _, f := range files {
		names = append(names, f.Name)
	}

	assert.Contains(t, names, "go.mod")
	assert.Contains(t, names, "main.go")
	assert.Contains(t, names, "Dockerfile")
}

func TestRenderSuccess(t *testing.T) {
	p := project.Project{
		Name:      "testapp",
		GoVersion: "1.26",
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
	p := project.Project{Name: "test"}

	// Pedir un archivo que no existe debe retornar un error
	_, err := templates.Render("test", p)

	assert.Error(t, err)
}

func TestRenderDockerfile(t *testing.T) {
	tests := []struct {
		name             string
		withDevcontainer bool
		expectedContains bool
	}{
		{
			name:             "WithDevcontainer",
			withDevcontainer: true,
			expectedContains: true,
		},
		{
			name:             "WithoutDevcontainer",
			withDevcontainer: false,
			expectedContains: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := project.Project{
				Name:             "testapp",
				GoVersion:        "1.26",
				WithDevcontainer: tt.withDevcontainer,
			}

			bytesResult, err := templates.Render("Dockerfile", p)
			assert.NoError(t, err)

			strContent := string(bytesResult)
			if tt.expectedContains {
				assert.Contains(t, strContent, "build-base")
				assert.Contains(t, strContent, "RUN go install golang.org/x/tools/gopls@latest")
			} else {
				assert.NotContains(t, strContent, "build-base")
				assert.NotContains(t, strContent, "RUN go install golang.org/x/tools/gopls@latest")
			}
		})
	}
}

func TestGetForbiddenFiles(t *testing.T) {
	forbidden := templates.GetForbiddenFiles(false)
	forbiddenWithOptionals := templates.GetForbiddenFiles(true)

	t.Run("contains all real template paths in lowercase", func(t *testing.T) {
		for _, spec := range templates.FileNames() {
			relPath := spec.Name
			if spec.Dir != "" {
				relPath = filepath.Join(spec.Dir, spec.Name)
			}
			lowerPath := strings.ToLower(relPath)
			assert.True(t, forbidden[lowerPath], "se esperaba que %q estuviera en la lista de prohibidos", lowerPath)
		}
	})

	tests := []struct {
		name       string
		forbidden  map[string]bool
		key        string
		expectTrue bool
	}{
		{"compose.yml variant", forbidden, "compose.yml", true},
		{"compose.yaml variant", forbidden, "compose.yaml", true},
		{"docker-compose.yml variant", forbidden, "docker-compose.yml", true},
		{"docker-compose.yaml variant", forbidden, "docker-compose.yaml", true},
		{"dockerfile variant", forbidden, "dockerfile", true},
		{"dockerfile.dev variant", forbidden, "dockerfile.dev", true},
		{"dockerfile.prod variant", forbidden, "dockerfile.prod", true},
		{"unrelated readme.md", forbidden, "readme.md", false},
		{"unrelated license", forbidden, "license", false},
		{"devcontainer.json in .devcontainer when true", forbiddenWithOptionals, ".devcontainer/devcontainer.json", true},
		{"devcontainer.json in .devcontainer when false", forbidden, ".devcontainer/devcontainer.json", false},
		{"devcontainer.json in root is NOT forbidden", forbiddenWithOptionals, "devcontainer.json", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectTrue {
				assert.True(t, tt.forbidden[tt.key])
			} else {
				assert.False(t, tt.forbidden[tt.key])
			}
		})
	}
}
