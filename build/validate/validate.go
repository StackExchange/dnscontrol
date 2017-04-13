package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/github"

	"golang.org/x/oauth2"
)

var status = ""

func appendErrorStatus(s string) {
	if status != "" {
		status += ", "
	}
	status += s
}

func main() {
	if err := checkGoFmt(); err != nil {
		fmt.Println(err)
		appendErrorStatus("needs gofmt")
	}
	if err := checkGoGenerate(); err != nil {
		fmt.Println(err)
		appendErrorStatus("needs go generate")
	}
	if status != "" {
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
	return fmt.Errorf("ERROR: The following files need to have gofmt run on them:\n%s", fList)
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

var client *github.Client
var commitish string

func init() {
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
		out, err := exec.Command("git", "rev-parse", tc+"^2").CombinedOutput()
		if err != nil {
			commitish = string(out)
		} else {
			log.Printf("Problem getting sha. Statuses will not be set")
		}
	}
	log.Println(commitish, "SHA")
}
