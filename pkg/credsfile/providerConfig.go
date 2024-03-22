// Package credsfile provides functions for reading and parsing the provider credentials json file.
// It cleans nonstandard json features (comments and trailing commas), as well as replaces environment variable placeholders with
// their environment variable equivalents. To reference an environment variable in your json file, simply use values in this format:
//
//	"key"="$ENV_VAR_NAME"
package credsfile

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/DisposaBoy/JsonConfigReader"
	"github.com/TomOnTime/utfutil"
	"github.com/google/shlex"
)

// LoadProviderConfigs will open or execute the specified file name, and parse its contents. It will replace environment variables it finds if any value matches $[A-Za-z_-0-9]+
func LoadProviderConfigs(fname string) (map[string]map[string]string, error) {
	var results = map[string]map[string]string{}

	var dat []byte
	var err error
	filesIsExecutable := strings.HasPrefix(fname, "!") || isExecutable(fname)

	if filesIsExecutable && !strings.HasSuffix(fname, ".json") {
		// file is executable and is not a .json (needed because in Windows WSL all files are executable).
		dat, err = executeCredsFile(strings.TrimPrefix(fname, "!"))
		if err != nil {
			return nil, err
		}
	} else {
		// no executable bit found nor marked as executable so read it in
		dat, err = readCredsFile(fname)
		if err != nil {
			return nil, err
		}
	}

	s := string(dat)
	r := JsonConfigReader.New(strings.NewReader(s))
	err = json.NewDecoder(r).Decode(&results)
	if err != nil {
		return nil, fmt.Errorf("failed parsing provider credentials file %v: %v", fname, err)
	}
	if err = replaceEnvVars(results); err != nil {
		return nil, err
	}

	// For backwards compatibility, insert NONE and BIND entries if
	// they do not exist. These are the only providers that previously
	// did not require entries in creds.json prior to v4.0.
	if _, ok := results["none"]; !ok {
		results["none"] = map[string]string{"TYPE": "NONE"}
	}
	if _, ok := results["bind"]; !ok {
		results["bind"] = map[string]string{"TYPE": "BIND"}
	}

	return results, nil
}

func isExecutable(filename string) bool {
	if stat, statErr := os.Stat(filename); statErr == nil {
		if mode := stat.Mode(); mode&0111 == 0111 {
			return true
		}
	}
	return false
}

func readCredsFile(filename string) ([]byte, error) {
	dat, err := utfutil.ReadFile(filename, utfutil.POSIX)
	if err != nil {
		// no creds file is ok. Bind requires nothing for example. Individual providers will error if things not found.
		if os.IsNotExist(err) {
			fmt.Printf("INFO: Config file %q does not exist. Skipping.\n", filename)
			return []byte{}, nil
		}
		return nil, fmt.Errorf("failed reading provider credentials file %v: %v", filename, err)
	}
	return dat, nil
}

func executeCredsFile(filename string) ([]byte, error) {
	cmd, err := shlex.Split(filename)
	if err != nil {
		return nil, err
	}
	command := cmd[0]

	// check if this is a file and not a command when there is no leading /
	if fileExists(command) && !strings.HasPrefix(command, "/") {
		command = strings.Join([]string{".", command}, string(filepath.Separator))
	}

	return exec.Command(command, cmd[1:]...).Output()
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, os.ErrNotExist)
}

func replaceEnvVars(m map[string]map[string]string) error {
	for _, keys := range m {
		for k, v := range keys {
			if strings.HasPrefix(v, "$") {
				env := v[1:]
				newVal := os.Getenv(env)
				keys[k] = newVal
			}
		}
	}
	return nil
}
