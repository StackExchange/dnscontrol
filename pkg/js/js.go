package js

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/pkg/printer"
	"github.com/StackExchange/dnscontrol/pkg/transform"
	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"              // load underscore js into vm by default
	_ "github.com/robertkrimen/otto/underscore" // required by otto
)

// currentDirectory is the current directory as used by require().
// This is used to emulate nodejs-style require() directory handling.
// If require("a/b/c.js") is called, any require() statement in c.js
// needs to be accessed relative to "a/b".  Therefore we
// track the currentDirectory (which is the current directory as
// far as require() is concerned, not the actual os.Getwd().
var currentDirectory string

// ExecuteJavascript accepts a javascript string and runs it, returning the resulting dnsConfig.
func ExecuteJavascript(file string, devMode bool) (*models.DNSConfig, error) {
	script, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Errorf("Reading js file %s: %s", file, err)
	}

	// Record the directory path leading up to this file.
	currentDirectory = filepath.Clean(filepath.Dir(file))

	vm := otto.New()

	vm.Set("require", require)
	vm.Set("REV", reverse)

	helperJs := GetHelpers(devMode)
	// run helper script to prime vm and initialize variables
	if _, err := vm.Run(helperJs); err != nil {
		return nil, err
	}

	// run user script
	if _, err := vm.Run(script); err != nil {
		return nil, err
	}

	// export conf as string and unmarshal
	value, err := vm.Run(`JSON.stringify(conf)`)
	if err != nil {
		return nil, err
	}
	str, err := value.ToString()
	if err != nil {
		return nil, err
	}
	conf := &models.DNSConfig{}
	if err = json.Unmarshal([]byte(str), conf); err != nil {
		return nil, err
	}
	return conf, nil
}

// GetHelpers returns the filename of helpers.js, or the esc'ed version.
func GetHelpers(devMode bool) string {
	return _escFSMustString(devMode, "/helpers.js")
}

func require(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) != 1 {
		throw(call.Otto, "require takes exactly one argument")
	}
	file := call.Argument(0).String()

	absFile := filepath.Clean(filepath.Join(currentDirectory, file))

	if strings.HasPrefix(file, ".") {
		file = absFile
	}

	// Record the old currentDirectory so that we can return there.
	currentDirectoryOld := currentDirectory
	// Record the directory path leading up to the file we're about to require.
	currentDirectory = filepath.Clean(filepath.Dir(absFile))

	printer.Debugf("requiring: %s\n", absFile)
	data, err := ioutil.ReadFile(file)

	if err != nil {
		throw(call.Otto, err.Error())
	}

	_, err = call.Otto.Run(string(data))
	if err != nil {
		throw(call.Otto, err.Error())
	}

	// Pop back to the old directory.
	currentDirectory = currentDirectoryOld

	return otto.TrueValue()
}

func throw(vm *otto.Otto, str string) {
	panic(vm.MakeCustomError("Error", str))
}

func reverse(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) != 1 {
		throw(call.Otto, "REV takes exactly one argument")
	}
	dom := call.Argument(0).String()
	rev, err := transform.ReverseDomainName(dom)
	if err != nil {
		throw(call.Otto, err.Error())
	}
	v, _ := otto.ToValue(rev)
	return v
}
