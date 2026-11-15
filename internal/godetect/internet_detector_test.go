package godetect

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeHTTPClient struct {
	responseBody string
	err          error
	delay        time.Duration
	statusCode   int
}

func (f fakeHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if f.err != nil {
		return nil, f.err
	}

	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
	}

	// statusCode zero-value => 200, para no rellenar el campo en cada caso de la tabla.
	statusCode := f.statusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(f.responseBody)),
	}, nil
}

func TestInternetDetector_Detect(t *testing.T) {

	validJSON := `[
		{
			"version": "go1.26.5",
			"stable": true,
			"files": []
		}
	]`

	emptyJSON := `[]`

	unexpectedFormatJSON := `[
		{
			"version": 123,
			"stable": true,
			"files": []
		}
	]`

	malformedJSON := `{"version": "go1.26.5" (corrupted json)`

	noPatchVersionJSON := `[
		{
			"version": "go1.26",
			"stable": true,
			"files": []
		}
	]`

	nonNumericVersionJSON := `[
		{
			"version": "gonext",
			"stable": true,
			"files": []
		}
	]`

	rcVersionJSON := `[
		{
			"version": "go1.27rc1",
			"stable": true,
			"files": []
		}
	]`

	tests := []struct {
		CaseName          string
		Body              string
		StatusCode        int
		expectedVersion   string
		ExpectError       bool
		ExpectedErrorMsg  string
		ExpectedAbsentMsg string
		Err               error
		Delay             time.Duration
		TimeoutOverride   time.Duration
	}{
		{
			CaseName:         "Normal Output",
			Body:             validJSON,
			expectedVersion:  "1.26.5",
			ExpectError:      false,
			ExpectedErrorMsg: "",
			Err:              nil,
			Delay:            0,
		},
		{
			CaseName:         "Timeout / Context Exceeded",
			Body:             validJSON,
			expectedVersion:  "",
			ExpectError:      true,
			ExpectedErrorMsg: "ocurrió un error en la respuesta",
			Err:              nil,
			TimeoutOverride:  10 * time.Millisecond,
			Delay:            50 * time.Millisecond,
		},
		{
			CaseName:         "Empty JSON response",
			Body:             emptyJSON,
			expectedVersion:  "",
			ExpectError:      true,
			ExpectedErrorMsg: "la respuesta JSON está vacía",
			Err:              nil,
			Delay:            0,
		},
		{
			CaseName:         "Unexpected JSON format",
			Body:             unexpectedFormatJSON,
			expectedVersion:  "",
			ExpectError:      true,
			ExpectedErrorMsg: "no se encontró el campo 'version' o no es un string",
			Err:              nil,
			Delay:            0,
		},
		{
			CaseName:         "Network error / Server down",
			Body:             "",
			expectedVersion:  "",
			ExpectError:      true,
			ExpectedErrorMsg: "ocurrió un error en la respuesta",
			Err:              errors.New("connection refused"),
			Delay:            0,
		},
		{
			CaseName:         "JSON Bad format",
			Body:             malformedJSON,
			expectedVersion:  "",
			ExpectError:      true,
			ExpectedErrorMsg: "error al parsear el JSON",
			Err:              nil,
			Delay:            0,
		},
		{
			CaseName:         "Non-200 status rejected (404)",
			Body:             validJSON,
			StatusCode:       http.StatusNotFound,
			expectedVersion:  "",
			ExpectError:      true,
			ExpectedErrorMsg: "status inesperado en la respuesta: 404 Not Found",
		},
		{
			CaseName:         "Non-200 status rejected (500)",
			Body:             validJSON,
			StatusCode:       http.StatusInternalServerError,
			expectedVersion:  "",
			ExpectError:      true,
			ExpectedErrorMsg: "status inesperado en la respuesta: 500 Internal Server Error",
		},
		{
			CaseName:         "Status error includes body snippet",
			Body:             `{"error": "mantenimiento programado"}`,
			StatusCode:       http.StatusServiceUnavailable,
			expectedVersion:  "",
			ExpectError:      true,
			ExpectedErrorMsg: "mantenimiento programado",
		},
		{
			CaseName:          "Status error with empty body omits snippet",
			Body:              "",
			StatusCode:        http.StatusForbidden,
			expectedVersion:   "",
			ExpectError:       true,
			ExpectedErrorMsg:  "status inesperado en la respuesta: 403",
			ExpectedAbsentMsg: "cuerpo:",
		},
		{
			CaseName:         "Non-numeric version rejected",
			Body:             nonNumericVersionJSON,
			expectedVersion:  "",
			ExpectError:      true,
			ExpectedErrorMsg: "formato válido",
		},
		{
			CaseName:         "Release candidate version rejected",
			Body:             rcVersionJSON,
			expectedVersion:  "",
			ExpectError:      true,
			ExpectedErrorMsg: "formato válido",
		},
		{
			CaseName:        "Version without patch",
			Body:            noPatchVersionJSON,
			expectedVersion: "1.26",
			ExpectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.CaseName, func(t *testing.T) {

			client := fakeHTTPClient{
				responseBody: tt.Body,
				err:          tt.Err,
				delay:        tt.Delay,
				statusCode:   tt.StatusCode,
			}

			id := NewInternetDetector(client)

			if tt.TimeoutOverride > 0 {
				id.timeout = tt.TimeoutOverride
			}

			version, err := id.Detect()

			if tt.ExpectError {
				assert.Error(t, err)
				assert.Empty(t, version)
				if err != nil {
					assert.Contains(t, err.Error(), tt.ExpectedErrorMsg)
					if tt.ExpectedAbsentMsg != "" {
						assert.NotContains(t, err.Error(), tt.ExpectedAbsentMsg)
					}
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedVersion, version)
			}
		})
	}
}
