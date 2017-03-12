package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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
