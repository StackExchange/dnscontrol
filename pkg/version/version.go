package version

import (
	"os/exec"
	"runtime/debug"
	"strings"
)

// Set by GoReleaser.
var version string

// VCSVersion retrieves the version information from git.
//
// If the current commit is untagged, the version string will show the last
// tag, followed by the number of commits since the tag, then the short
// hash of the current commit.
//
// If the tree is dirty, "-dirty" is appended.
func VCSVersion() string {
	cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	v, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	ver := strings.TrimSpace(string(v))
	return ver
}

// Version returns either the tag set by GoReleaser, or the version information
// from Git.
func Version() string {
	if version != "" {
		return version
	}
	bi, ok := debug.ReadBuildInfo()
	if !ok ||
		// When running with "go run main.go" no module information is available
		bi.Main.Version == "" ||
		// Go gives no commit information if not on a tag
		bi.Main.Version == "(devel)" {
		return VCSVersion()
	}
	return bi.Main.Version
}
