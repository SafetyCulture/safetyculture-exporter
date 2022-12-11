package update_test

import (
	"testing"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/update"
)

func Test_versionGreaterThanOrEqual(t *testing.T) {
	type args struct {
		v string
		w string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "when smaller",
			args: args{v: "0.0.0", w: "0.0.1"},
			want: false,
		},
		{
			name: "when greater",
			args: args{v: "0.0.2", w: "0.0.1"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := update.VersionGreaterThanOrEqual(tt.args.v, tt.args.w); got != tt.want {
				t.Errorf("versionGreaterThanOrEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
