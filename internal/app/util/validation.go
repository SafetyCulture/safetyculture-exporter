package util

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var lgr *zap.SugaredLogger

func Check(err error, msg string) {
	if lgr == nil {
		lgr = GetLogger()
	}

	if err != nil {
		lgr.Fatal(errors.Wrapf(err, msg))
	}
}
