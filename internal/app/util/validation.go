package util

import (
	"net/http"

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

// CheckFeedError - checks the Feed for errors, except for 403's
func CheckFeedError(err error, msg string) {
	if err == nil {
		return
	}

	if lgr == nil {
		lgr = GetLogger()
	}

	switch e := err.(type) {
	case HttpError:
		if e.StatusCode == http.StatusForbidden {
			lgr.Error(errors.Wrapf(err, msg))
			return
		}
	}
	lgr.Fatal(errors.Wrapf(err, msg))
}
