package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

var sha = flag.String("sha", "", "SHA of current commit")

var goos = flag.String("os", "", "OS to build (linux, windows, or darwin) Defaults to all.")

func main() {
	flag.Parse()
	flags := fmt.Sprintf(`-s -w -X "main.version=%s"`, getVersion())
	pkg := "github.com/StackExchange/dnscontrol/v4"

	build := func(out, goos string) {
		log.Printf("Building %s", out)
		cmd := exec.Command("go", "build", "-o", out, "-ldflags", flags, pkg)
		os.Setenv("GOOS", goos)
		os.Setenv("GO111MODULE", "on")
		os.Setenv("CGO_ENABLED", "0")
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		// For now just build for amd64 everywhere
		if os.Getenv("GOARCH") == "" {
			os.Setenv("GOARCH", "amd64")
		}

		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, env := range []struct {
		binary, goos string
	}{
		{"dnscontrol-Linux", "linux"},
		{"dnscontrol.exe", "windows"},
		{"dnscontrol-Darwin", "darwin"},
	} {
		if *goos == "" || *goos == env.goos {
			build(env.binary, env.goos)
		}
	}
}

func getVersion() string {
	if *sha != "" {
		return *sha
	}
	// check teamcity build version
	if v := os.Getenv("BUILD_VCS_NUMBER"); v != "" {
		return v
	}
	// check git
	cmd := exec.Command("git", "rev-parse", "HEAD")
	v, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	ver := strings.TrimSpace(string(v))
	// see if dirty
	cmd = exec.Command("git", "diff-index", "--quiet", "HEAD", "--")
	err = cmd.Run()
	// exit status 1 indicates dirty tree
	if err != nil {
		if err.Error() == "exit status 1" {
			ver += "[dirty]"
		} else {
			log.Printf("!%s!", err.Error())
		}
	}
	return ver
}
