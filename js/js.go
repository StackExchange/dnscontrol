package js

import (
	"encoding/json"

	"github.com/StackExchange/dnscontrol/models"

	"github.com/robertkrimen/otto"
	//load underscore js into vm by default
	_ "github.com/robertkrimen/otto/underscore"
)

//ExecuteJavascript accepts a javascript string and runs it, returning the resulting dnsConfig.
func ExecuteJavascript(script string, devMode bool) (*models.DNSConfig, error) {
	vm := otto.New()

	helperJs := GetHelpers(devMode)
	// run helper script to prime vm and initialize variables
	if _, err := vm.Run(string(helperJs)); err != nil {
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

func GetHelpers(devMode bool) string {
	return FSMustString(devMode, "/helpers.js")
}
