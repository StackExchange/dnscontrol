// Package config provides functions for reading and parsing the provider credentials json file.
// It cleans nonstandard json features (comments and trailing commas), as well as replaces environment variable placeholders with
// their environment variable equivalents. To reference an environment variable in your json file, simply use values in this format:
//    "key"="$ENV_VAR_NAME"
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/DisposaBoy/JsonConfigReader"
	"github.com/TomOnTime/utfutil"
)

// LoadProviderConfigs will open the specified file name, and parse its contents. It will replace environment variables it finds if any value matches $[A-Za-z_-0-9]+
func LoadProviderConfigs(fname string) (map[string]map[string]string, error) {
	var results = map[string]map[string]string{}
	dat, err := utfutil.ReadFile(fname, utfutil.POSIX)
	if err != nil {
		// no creds file is ok. Bind requires nothing for example. Individual providers will error if things not found.
		if os.IsNotExist(err) {
			fmt.Printf("INFO: Config file %q does not exist. Skipping.\n", fname)
			return results, nil
		}
		return nil, fmt.Errorf("failed reading provider credentials file %v: %v", fname, err)
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
	return results, nil
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
