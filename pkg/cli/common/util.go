package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// CheckErr prints a user friendly error to STDERR and exits with a non-zero
// exit code. This function is used as a wrapper for the set of steps that comprise
// the execution of a cobra command. It is the common exit point used by
// all cobra `Run` handlers. This convention, in combination with the fact that
// cobra commands only use `Run` handlers, but not `RunE` handlers (i.e. they
// don't return errors back to the cobra framework), allows for better control
// over where and how errors are handled.
func CheckErr(err error) {
	if err == nil {
		return
	}

	msg := err.Error()
	if !strings.HasPrefix(msg, "error: ") {
		msg = fmt.Sprintf("error: %s", msg)
	}
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(-1)
}

// UnpackLabelArgs converts a list of strings into a map. This helper function can be used
// to unpack command line arguments used to supplye dictionary values, e.g.:
//
//   --label foo:bar --label fan: --label fin
//
// can be collected as an array of strings:
//
//   ["foo:bar", "fan:", "fin"]
//
// and then unpacked with this function into a corresponding map:
//
//   {"foo": "bar", "fan": "", "fin":""}
//
func UnpackLabelArgs(labelArgs []string, labels map[string]string) {
	for _, l := range labelArgs {
		var k, v string
		l = strings.TrimSpace(l)
		s := strings.Split(l, ":")
		if len(s) > 1 {
			k = strings.TrimSpace(s[0])
			v = strings.TrimSpace(strings.Join(s[1:], ":"))
		} else if len(s) == 1 {
			k = strings.TrimSpace(s[0])
			v = ""
		} else {
			k = l
			v = ""
		}
		labels[k] = v
	}
}

// LoadFileIntoVar loads the entire contents of a file into the supplied string variable
func LoadFileIntoVar(filePath string, destContent *string) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cannot read file %s: %w", filePath, err)
	}

	*destContent = string(content)

	return nil
}
