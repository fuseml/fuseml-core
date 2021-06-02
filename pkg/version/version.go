package version

import (
	"encoding/json"
	"fmt"
	"runtime"
)

var (
	// Version is the CLI version
	Version string = "unknown"

	// GitCommit is the git commit ID
	GitCommit string = "unknown"

	// BuildDate is the date when the binary was built
	BuildDate string = "unknown"
)

// Info contains versioning information.
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
	Compiler  string `json:"compiler"`
	Platform  string `json:"platform"`
}

// GetInfo returns versioning information.
func GetInfo() *Info {
	return &Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// GetInfoStr returns versioning information in a string format.
func GetInfoStr() string {
	i := GetInfo()

	jsonVer, _ := json.Marshal(i)

	return string(jsonVer)
}
