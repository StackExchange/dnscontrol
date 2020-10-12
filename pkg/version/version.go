package version

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"time"
)

// Version management. Goals:
// 1. Someone who just does "go get" has at least some information.
// 2. If built with build/build.go, more specific build information gets put in.
// Update the number here manually each release, so at least we have a range for go-get people.
var (
	SHA       = ""
	Version   = "3.3.0"
	BuildTime = ""
)

var versionCache string

// VersionString returns the version banner.
func VersionString() string {
	if versionCache != "" {
		return versionCache
	}

	var version string
	if SHA != "" {
		version = fmt.Sprintf("%s (%s)", Version, SHA)
	} else {
		version = fmt.Sprintf("%s-dev", Version) // no SHA. '0.x.y-dev' indicates it is run from source without build script.
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
