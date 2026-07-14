package godetect

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeStrategy struct {
	version string
	err     error
}

func (s fakeStrategy) Detect() (string, error) {
	return s.version, s.err
}

func TestDetect(t *testing.T) {

	var (
		errA = errors.New("error en estrategia A")
		errB = errors.New("error en estrategia B")
	)

	tests := []struct {
		CaseName		string
		inputStrategies	[]VersionStrategy
		expectedVersion string
		expectedErrors    []error
	}{
		{
			CaseName:        	"The first strategy works",
			inputStrategies: []VersionStrategy{
				fakeStrategy{version: "1.21.0", err: nil},
			},
			expectedVersion:	"1.21.0",
			expectedErrors:		nil,
		},
		{
			CaseName:        	"The first strategy fails; the second one finds the version",
			inputStrategies: []VersionStrategy{
				fakeStrategy{version: "", err: errors.New("permiso denegado")},
				fakeStrategy{version: "1.22.0", err: nil},
			},
			expectedVersion:	"1.22.0",
			expectedErrors:		nil,
		},
		{
			CaseName: 			"All strategies fail",
			inputStrategies: 	[]VersionStrategy{
				fakeStrategy{version: "", err: errA},
				fakeStrategy{version: "", err: errB},
			},
			expectedVersion: 	"",
			expectedErrors:     []error{errA, errB},
		},
	}

	for _, tt := range tests {
		t.Run(tt.CaseName, func(t *testing.T) {
			version, err := Detect(tt.inputStrategies)

			assert.Equal(t, tt.expectedVersion, version)

			if len(tt.expectedErrors) > 0 {
				assert.Error(t, err)

				for _, expectedErr := range tt.expectedErrors {
					assert.ErrorIs(t, err, expectedErr, "El error devuelto debe contener: %v", expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}