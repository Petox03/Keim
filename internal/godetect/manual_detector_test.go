package godetect_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"keim/internal/godetect"
)

func TestManualDetector_Detect(t *testing.T) {
	tests := []struct {
		CaseName          string
		inputVersion      string
		expectedVersion   string
		ExpectedErrSuffix string
	}{
		{
			CaseName:          "Explicit version provided",
			inputVersion:      "1.26.2",
			expectedVersion:   "1.26.2",
			ExpectedErrSuffix: "",
		},
		{
			CaseName:          "Empty version returns error (stdin unsupported)",
			inputVersion:      "",
			expectedVersion:   "",
			ExpectedErrSuffix: "la versión de Go no fue especificada",
		},
	}

	for _, tt := range tests {
		t.Run(tt.CaseName, func(t *testing.T) {
			md := godetect.NewManualDetector(tt.inputVersion)

			version, err := md.Detect()

			if tt.ExpectedErrSuffix == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedVersion, version)
			} else {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.ExpectedErrSuffix)
			}
		})
	}
}
