package version

import "runtime/debug"

// Set by GoReleaser
var version string

func Version() string {
	if version != "" {
		return version
	}
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	return bi.Main.Version
}
