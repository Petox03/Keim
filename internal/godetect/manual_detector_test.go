package godetect_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"keim/internal/godetect"
)

func TestManualDetector_Detect(t *testing.T) {
	tests := []struct {
		CaseName        string
		inputVersion    string
		stdin           string
		wantVersion     string
		wantErr         bool
		wantErrContains []string
		wantStdoutEmpty bool
	}{
		{
			CaseName:        "Explicit version with patch",
			inputVersion:    "1.26.2",
			stdin:           "",
			wantVersion:     "1.26.2",
			wantStdoutEmpty: true,
		},
		{
			CaseName:        "Explicit version without patch",
			inputVersion:    "1.26",
			stdin:           "",
			wantVersion:     "1.26",
			wantStdoutEmpty: true,
		},
		{
			CaseName:        "Empty version, valid input on first try",
			inputVersion:    "",
			stdin:           "1.27.0\n",
			wantVersion:     "1.27.0",
			wantStdoutEmpty: false,
		},
		{
			CaseName:        "Empty version, invalid then valid",
			inputVersion:    "",
			stdin:           "abc\n1.27.0\n",
			wantVersion:     "1.27.0",
			wantStdoutEmpty: false,
		},
		{
			CaseName:        "Empty version, retries exhausted",
			inputVersion:    "",
			stdin:           "x\ny\nz\n",
			wantVersion:     "",
			wantErr:         true,
			wantErrContains: []string{"manual detector:", "reintentos"},
			wantStdoutEmpty: false,
		},
		{
			CaseName:        "Empty version, immediate EOF",
			inputVersion:    "",
			stdin:           "",
			wantVersion:     "",
			wantErr:         true,
			wantErrContains: []string{"manual detector:", "reintentos"},
			wantStdoutEmpty: false,
		},
		{
			CaseName:        "Regex rejects v-prefix",
			inputVersion:    "",
			stdin:           "v1.26.2\n1.26.2\n",
			wantVersion:     "1.26.2",
			wantStdoutEmpty: false,
		},
		{
			CaseName:        "Regex rejects single component",
			inputVersion:    "",
			stdin:           "1\n1.26.2\n",
			wantVersion:     "1.26.2",
			wantStdoutEmpty: false,
		},
		{
			CaseName:        "Regex rejects four components",
			inputVersion:    "",
			stdin:           "1.2.3.4\n1.26.2\n",
			wantVersion:     "1.26.2",
			wantStdoutEmpty: false,
		},
		{
			CaseName:        "Regex rejects trailing dot",
			inputVersion:    "",
			stdin:           "1.26.\n1.26.2\n",
			wantVersion:     "1.26.2",
			wantStdoutEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.CaseName, func(t *testing.T) {
			var stdout bytes.Buffer
			md := godetect.NewManualDetector(tt.inputVersion, bytes.NewReader([]byte(tt.stdin)), &stdout)

			version, err := md.Detect()

			if tt.wantErr {
				assert.Error(t, err)
				for _, sub := range tt.wantErrContains {
					assert.ErrorContains(t, err, sub)
				}
				assert.Empty(t, version)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantVersion, version)
			}

			if tt.wantStdoutEmpty {
				assert.Empty(t, stdout.String())
			} else {
				assert.NotEmpty(t, stdout.String())
			}
		})
	}
}
