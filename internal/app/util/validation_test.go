// Copyright (c) 2018 SafetyCulture Pty Ltd. All Rights Reserved.

package util_test

import (
	"testing"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{
			name: "Should not panic if there is no error",
			err:  nil,
			msg:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			util.Check(tt.err, tt.msg)
		})
	}
}
