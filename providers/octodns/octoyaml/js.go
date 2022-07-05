package octoyaml

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/transform"

	"github.com/robertkrimen/otto"
	// load underscore js into vm by default

	_ "github.com/robertkrimen/otto/underscore" // required by otto
)

// ExecuteJavascript accepts a javascript string and runs it, returning the resulting dnsConfig.
func ExecuteJavascript(script string, devMode bool) (*models.DNSConfig, error) {
	vm := otto.New()

	vm.Set("require", require)
	vm.Set("REV", reverse)

	helperJs := GetHelpers(true)
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
	d, err := ioutil.ReadFile("../pkg/js/helpers.js")
	if err != nil {
		panic(err)
	}
	return string(d)
}

func require(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) != 1 {
		throw(call.Otto, "require takes exactly one argument")
	}
	file := call.Argument(0).String()
	fmt.Printf("requiring: %s\n", file)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		throw(call.Otto, err.Error())
	}
	_, err = call.Otto.Run(string(data))
	if err != nil {
		throw(call.Otto, err.Error())
	}
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
