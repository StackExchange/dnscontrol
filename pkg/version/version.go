package version

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"time"
)

// NOTE: main() updates these.
var (
	SHA       = ""
	Semver    = ""
	BuildTime = ""
)

var versionCache string

// Banner returns the version banner.
func Banner() string {
	if versionCache != "" {
		return versionCache
	}

	var version string
	if SHA != "" {
		version = fmt.Sprintf("%s (%s)", Semver, SHA)
	} else {
		version = fmt.Sprintf("%s-dev", Semver) // no SHA. '0.x.y-dev' indicates it is run from source without build script.
	}
	if info, ok := debug.ReadBuildInfo(); !ok && info == nil {
		version += " (non-modules)"
	}
	if BuildTime != "" {
		i, err := strconv.ParseInt(BuildTime, 10, 64)
		if err == nil {
			tm := time.Unix(i, 0)
			version += fmt.Sprintf(" built %s", tm.Format(time.RFC822))
		}
	}

	versionCache = version
	return version
}
