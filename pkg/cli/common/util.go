package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// KeyValueArgs is used with key-value command line arguments
type KeyValueArgs struct {
	Packed   []string
	Unpacked map[string]string
}

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

// Unpack converts a list of strings into a map. This helper function can be used
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
func (args *KeyValueArgs) Unpack() {
	args.Unpacked = make(map[string]string)
	for _, l := range args.Packed {
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
		args.Unpacked[k] = v
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

// WriteConfigFile writes new content of the config file.
// If the file does not exist, it is created at default location
// TODO temporary solution until upstream https://github.com/spf13/viper/issues/433 is fixed
func WriteConfigFile() error {
	cf := viper.ConfigFileUsed()

	if cf == "" {
		fullname := ConfigFileName + "." + ConfigFileType
		if dirname, err := os.UserHomeDir(); err == nil {
			cf = filepath.Join(dirname, ConfigHomeSubdir, ConfigFuseMLSubdir, fullname)
		}
		if cf == "" {
			return errors.New("Failed to acquire config directory name")
		}
		configDirPath := filepath.Dir(cf)
		if err := os.MkdirAll(configDirPath, os.ModePerm); err != nil {
			return err
		}

		fmt.Printf("FuseML configuration file created at %s\n", cf)
	}

	if err := viper.WriteConfigAs(cf); err != nil {
		return err
	}
	return nil
}

// StringInSlice verifies if a string slice contains a string value
func StringInSlice(s string, slice []string) bool {
	for _, v := range slice {
		if s == v {
			return true
		}
	}
	return false
}

// ValidateEnumArgument is used to validate command line arguments that can take a limited set of values
func ValidateEnumArgument(argName, argValue string, values []string) error {
	if !StringInSlice(argValue, values) {
		return fmt.Errorf("%s must be one of: %s", argName, strings.Join(values, ", "))
	}
	return nil
}
