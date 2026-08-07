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

	return &http.Response{
		StatusCode: 200,
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

	tests := []struct {
		CaseName         string
		Body             string
		expectedVersion  string
		ExpectError      bool
		ExpectedErrorMsg string
		Err              error
		Delay            time.Duration
		TimeoutOverride  time.Duration
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
	}

	for _, tt := range tests {
		t.Run(tt.CaseName, func(t *testing.T) {

			client := fakeHTTPClient{
				responseBody: tt.Body,
				err:          tt.Err,
				delay:        tt.Delay,
			}

			id := NewInternetDetector(client)

			if tt.TimeoutOverride > 0 {
				id.timeout = tt.TimeoutOverride
			}

			version, err := id.Detect()

			if tt.ExpectError {
				assert.Error(t, err)
				assert.Empty(t, version)
				assert.Contains(t, err.Error(), tt.ExpectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedVersion, version)
			}
		})
	}
}
