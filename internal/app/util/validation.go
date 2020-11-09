package util

import (
	"github.com/pkg/errors"
)

func Check(err error, msg string) {
	logger := GetLogger()

	if err != nil {
		logger.Fatal(errors.Wrapf(err, msg))
	}
}
