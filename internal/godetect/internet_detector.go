package godetect

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type InternetDetector struct {
	client  HTTPClient
	timeout time.Duration
}

func NewInternetDetector(c HTTPClient) *InternetDetector {
	if c == nil {
		c = &http.Client{}
	}

	return &InternetDetector{
		client:  c,
		timeout: 1 * time.Second,
	}
}

func (id *InternetDetector) Detect() (version string, retErr error) {
	ctx, cancel := context.WithTimeout(context.Background(), id.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://go.dev/dl/?mode=json", nil)
	if err != nil {
		return "", fmt.Errorf("ocurrió un error al crear la petición: %w", err)
	}

	res, err := id.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ocurrió un error en la respuesta: %w", err)
	}
	defer func() {
		if cerr := res.Body.Close(); cerr != nil && retErr == nil {
			retErr = fmt.Errorf("error al cerrar el cuerpo de la respuesta: %w", cerr)
		}
	}()

	if res.StatusCode != http.StatusOK {
		status := fmt.Sprintf("%d %s", res.StatusCode, http.StatusText(res.StatusCode))
		snippet, rerr := io.ReadAll(io.LimitReader(res.Body, 512))
		if rerr != nil {
			return "", fmt.Errorf("status inesperado en la respuesta: %s (no se pudo leer el cuerpo: %w)", status, rerr)
		}
		if cuerpo := strings.Join(strings.Fields(string(snippet)), " "); cuerpo != "" {
			return "", fmt.Errorf("status inesperado en la respuesta: %s (cuerpo: %s)", status, cuerpo)
		}
		return "", fmt.Errorf("status inesperado en la respuesta: %s", status)
	}

	var result []map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error al parsear el JSON; %w", err)
	}

	if len(result) == 0 {
		return "", fmt.Errorf("la respuesta JSON está vacía")
	}

	rawVersion, ok := result[0]["version"].(string)
	if !ok {
		return "", fmt.Errorf("no se encontró el campo 'version' o no es un string")
	}

	version = strings.TrimPrefix(rawVersion, "go")

	if !versionRegex.MatchString(version) {
		return "", fmt.Errorf("la versión recibida no tiene un formato válido: %s", rawVersion)
	}

	return version, nil
}
