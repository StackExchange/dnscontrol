package main

// ReadHints returns the "hints" configuration as a data structure.  (Right now
// it returns hardcoded constants.  In the future it should read a YAML file.)
func GetHints() ([]string, TypeCatalog) {

	var l []string
	var cat TypeCatalog = TypeCatalog{}

	addType := func(name string, token string, fields []Field) {
		l = append(l, name)
		n := RTypeConfig{}
		if token != "" {
			n.Token = token
		}
		if fields != nil {
			n.Fields = fields
		}
		cat[name] = n
	}
	setNoLabel := func(name string) {
		n := cat[name]
		{
			n.NoLabel = true
		}
		cat[name] = n
	}
	setTTL1 := func(name string) {
		n := cat[name]
		{
			n.TTL1 = true
		}
		cat[name] = n
	}

	addType("A", "", nil)

	addType("MX", "", nil)

	addType("SRV", "",
		[]Field{
			{Name: "Priority", Tags: SloppyParseTags(`json:"priority"`)},
			{Name: "Weight", Tags: SloppyParseTags(`json:"weight"`)},
			{Name: "Port", Tags: SloppyParseTags(`json:"port"`)},
			{Name: "Target", Tags: SloppyParseTags(`json:"target" dns:"domain-name"`)},
		},
	)

	addType("CFSINGLEREDIRECT", "CF_SINGLE_REDIRECT", nil)
	setNoLabel("CFSINGLEREDIRECT")
	setTTL1("CFSINGLEREDIRECT")

	return l, cat
}
