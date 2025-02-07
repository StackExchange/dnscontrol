package main

// ReadHints returns the "hints" configuration as a data structure.  (Right now
// it returns hardcoded constants.  In the future it should read a YAML file.)
func GetHints() Catalog {

	// Some
	return Catalog{

		"A": RTypeConfig{},

		"MX": RTypeConfig{},

		"SRV": RTypeConfig{
			Fields: []Field{
				{Name: "Target", Type: "int16"},
			},
		},

		"CFSINGLEREDIRECT": RTypeConfig{
			Token: "CF_SINGLE_REDIRECT",
		},
	}
}
