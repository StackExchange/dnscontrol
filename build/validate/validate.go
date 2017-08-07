package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/go-github/github"

	"golang.org/x/oauth2"
)

func main() {
	failed := false

	run := func(ctx string, preStatus string, goodStatus string, f func() error) {
		setStatus(stPending, preStatus, ctx)
		if err := f(); err != nil {
			fmt.Println(err)
			setStatus(stError, err.Error(), ctx)
			failed = true
		} else {
			setStatus(stSuccess, goodStatus, ctx)
		}
	}

	run("gofmt", "Checking gofmt", "gofmt ok", checkGoFmt)
	run("gogen", "Checking go generate", "go generate ok", checkGoGenerate)
	run("", "Checking for direct console output", "", checkDirectOutput)

	if failed {
		os.Exit(1)
	}
}

func checkGoFmt() error {
	cmd := exec.Command("gofmt", "-s", "-l", ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	if len(out) == 0 {
		return nil
	}
	files := strings.Split(string(out), "\n")
	fList := ""
	for _, f := range files {
		if strings.HasPrefix(f, "vendor") {
			continue
		}
		if fList != "" {
			fList += "\n"
		}
		fList += f
	}
	if fList == "" {
		return nil
	}
	return fmt.Errorf("The following files need to have gofmt run on them:\n%s", fList)
}

func checkGoGenerate() error {
	cmd := exec.Command("go", "generate")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	modified, err := getModifiedFiles()
	if err != nil {
		return err
	}
	if len(modified) != 0 {
		return fmt.Errorf("ERROR: The following files are modified after go generate:\n%s", strings.Join(modified, "\n"))
	}
	return nil
}

func getModifiedFiles() ([]string, error) {
	cmd := exec.Command("git", strings.Split("diff --name-only", " ")...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, nil
	}
	return strings.Split(string(out), "\n"), nil
}

func checkDirectOutput() error {
	var files []string
	var err error
	var badStrings = []string{
		"fmt.Print",
		"log.Print",
	}
	var rd func(string)
	rd = func(path string) {
		if path == ".git" || path == "vendor" || path == "build" {
			return
		}
		var fis []os.FileInfo
		fis, err = ioutil.ReadDir(path)
		if err != nil {
			return
		}
		for _, fi := range fis {
			n := fi.Name()
			fp := filepath.Join(path, n)
			if fi.IsDir() {
				rd(fp)
			} else if filepath.Ext(n) == ".go" {
				var dat []byte
				dat, err = ioutil.ReadFile(fp)
				if err != nil {
					return
				}
				for _, bs := range badStrings {
					if strings.Contains(string(dat), bs) {
						files = append(files, fp)
						break
					}
				}
			}
		}
	}
	rd(".")
	if err != nil {
		return err
	}
	err = fmt.Errorf("The following files contain printing direct to console:\n%s", strings.Join(files, "\n"))
	// eat error for now, just print as info until we clean it
	fmt.Println(err.Error())
	return nil
}

const (
	stPending = "pending"
	stSuccess = "success"
	stError   = "error"
)

func setStatus(status string, desc string, ctx string) {
	if commitish == "" || ctx == "" {
		return
	}
	client.Repositories.CreateStatus(context.Background(), "StackExchange", "dnscontrol", commitish, &github.RepoStatus{
		Context:     &ctx,
		Description: &desc,
		State:       &status,
	})
}

var client *github.Client
var commitish string

func init() {
	// not intended for security, just minimal obfuscation.
	key, _ := base64.StdEncoding.DecodeString("qIOy76aRcXcxm3vb82tvZqW6JoYnpncgVKx7qej1y+4=")
	iv, _ := base64.StdEncoding.DecodeString("okRtW8z6Mx04Y9yMk1cb5w==")
	garb, _ := base64.StdEncoding.DecodeString("ut8AtS6re1g7m/onk0ciIq7OxNOdZ/tsQ5ay6OfxKcARnBGY0bQ+pA==")
	c, _ := aes.NewCipher(key)
	d := cipher.NewCFBDecrypter(c, iv)
	t := make([]byte, len(garb))
	d.XORKeyStream(t, garb)
	hc := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: string(t)}))
	client = github.NewClient(hc)

	//get current version if in travis build
	if tc := os.Getenv("TRAVIS_COMMIT"); tc != "" {
		commitish = tc
	}
}
