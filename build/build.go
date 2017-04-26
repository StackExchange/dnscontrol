package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

var sha = flag.String("sha", "", "SHA of current commit")

func main() {
	flag.Parse()
	flags := fmt.Sprintf(`-s -w -X main.SHA="%s" -X main.BuildTime=%d`, getVersion(), time.Now().Unix())
	pkg := "github.com/StackExchange/dnscontrol"

	build := func(out, goos string) {
		log.Printf("Building %s", out)
		cmd := exec.Command("go", "build", "-o", out, "-ldflags", flags, pkg)
		os.Setenv("GOOS", goos)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}

	build("dnscontrol-Linux", "linux")
	build("dnscontrol.exe", "windows")
	build("dnscontrol-Darwin", "darwin")
}

func getVersion() string {
	if *sha != "" {
		return *sha
	}
	//check teamcity build version
	if v := os.Getenv("BUILD_VCS_NUMBER"); v != "" {
		return v
	}
	//check git
	cmd := exec.Command("git", "rev-parse", "HEAD")
	v, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	ver := strings.TrimSpace(string(v))
	//see if dirty
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
