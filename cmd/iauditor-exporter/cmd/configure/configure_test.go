package configure_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/SafetyCulture/iauditor-exporter/cmd/iauditor-exporter/cmd"
)

func TestCommandConfigure_should_not_throw_error(t *testing.T) {
	b := bytes.NewBufferString("")
	cmd.RootCmd.SetOut(b)
	cmd.RootCmd.SetArgs([]string{"configure"})
	cmd.Execute()
	_, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
}
