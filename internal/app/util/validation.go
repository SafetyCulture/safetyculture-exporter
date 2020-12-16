package util

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var lgr *zap.SugaredLogger

// Check the error and exit the process when not nil
func Check(err error, msg string) {
	if lgr == nil {
		lgr = GetLogger()
	}

	if err != nil {
		lgr.Fatal(errors.Wrapf(err, msg))
	}
}
