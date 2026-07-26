package godetect

import (
	"errors"
)

func Detect(strategies []VersionStrategy) (string, error) {
	errorList := make([]error, 0, len(strategies))

	for _, strategy := range strategies {
		version, err := strategy.Detect()
		if err == nil {
			return version, nil
		}
		errorList = append(errorList, err)
	}

	return "", errors.Join(errorList...)
}
