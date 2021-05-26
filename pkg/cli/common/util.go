package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// CheckErr prints a user friendly error to STDERR and exits with a non-zero
// exit code.
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

func LoadFileIntoVar(filePath string, destContent *string) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cannot read file %s: %w", filePath, err)
	}

	*destContent = string(content)

	return nil
}
