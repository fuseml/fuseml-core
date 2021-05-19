package common

import (
	"fmt"
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
	fmt.Fprint(os.Stderr, msg)
	os.Exit(-1)
}
