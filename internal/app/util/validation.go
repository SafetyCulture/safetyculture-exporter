package util

import (
	"github.com/pkg/errors"
)

// Check the error and exit the process when not nil
func Check(err error, msg string) {
	logger := GetLogger()

	if err != nil {
		logger.Fatal(errors.Wrapf(err, msg))
	}
}
