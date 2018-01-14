/**
 * Copyright 2016 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"html/template"
	"os"

	"flag"
	"github.com/softlayer/softlayer-go/sl"
)

var versionfile = fmt.Sprintf(`%s

%s

package sl

import "fmt"

type VersionInfo struct {
	Major  int
	Minor  int
	Patch  int
	Pre    string
}

var Version = VersionInfo {
	Major:  {{.Major}},
	Minor:  {{.Minor}},
	Patch:  {{.Patch}},
	Pre:    "{{.Pre}}",
}

func (v VersionInfo) String() string {
	result := fmt.Sprintf("v%%d.%%d.%%d", v.Major, v.Minor, v.Patch)

	if v.Pre != "" {
		result += fmt.Sprintf("-%%s", v.Pre)
	}

	return result
}

`, license, codegenWarning)

const (
	bumpUsage   = "specify one of major|minor|patch to bump that specifier"
	prerelUsage = "optional prerelease stamp (e.g. alpha, beta, rc.1"
)

func version() {
	var bump, prerelease string

	flagset := flag.NewFlagSet(os.Args[1], flag.ExitOnError)

	flagset.StringVar(&bump, "bump", "", bumpUsage)
	flagset.StringVar(&bump, "b", "", bumpUsage+" (shorthand)")

	flagset.StringVar(&prerelease, "prerelease", "", prerelUsage)
	flagset.StringVar(&prerelease, "p", "", prerelUsage+" (shorthand)")

	flagset.Parse(os.Args[2:])

	v := sl.Version

	switch bump {
	case "major":
		v.Major++
		v.Minor = 0
		v.Patch = 0
		v.Pre = prerelease
	case "minor":
		v.Minor++
		v.Patch = 0
		v.Pre = prerelease
	case "patch":
		v.Patch++
		v.Pre = prerelease
	case "":
	default:
		bail(fmt.Errorf("Invalid value for bump: %s", bump))
	}

	writeVersionFile(v)

	fmt.Println(v)
}

func writeVersionFile(v sl.VersionInfo) {
	// Generate source
	var buf bytes.Buffer
	t := template.New("version")
	template.Must(t.Parse(versionfile)).Execute(&buf, v)

	//fmt.Println(string(buf.Bytes()))

	// format go file
	pretty, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// write go file
	f, err := os.Create("sl/version.go")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	fmt.Fprintf(f, "%s", pretty)
}
