package httpapi

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultBackoff(t *testing.T) {
	type args struct {
		min        time.Duration
		max        time.Duration
		attemptNum int
		resp       *http.Response
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{
			name: "When Response is nil and attempt is 0",
			args: args{
				min:        0,
				max:        0,
				attemptNum: 0,
				resp:       nil,
			},
			want: 0,
		},
		{
			name: "When Response is nil sleep > max",
			args: args{
				min:        1,
				max:        1,
				attemptNum: 1,
				resp:       nil,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, DefaultBackoff(tt.args.min, tt.args.max, tt.args.attemptNum, tt.args.resp), "DefaultBackoff(%v, %v, %v, %v)", tt.args.min, tt.args.max, tt.args.attemptNum, tt.args.resp)
		})
	}
}
