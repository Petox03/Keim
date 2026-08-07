package godetect

import (
	"context"
	"encoding/json"
	"fmt"
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

func (id *InternetDetector) Detect() (string, error) {

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
	defer res.Body.Close()

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

	version := strings.TrimPrefix(rawVersion, "go")

	return version, nil

}
