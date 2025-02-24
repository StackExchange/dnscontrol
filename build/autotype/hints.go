package main

// GetHints returns the "hints" configuration as a data structure.  (Right now
// it returns hardcoded constants.  In the future it should read a configuration file.)
func GetHints() ([]string, TypeCatalog) {

	var l []string
	var cat = TypeCatalog{}

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
	setIsBuilder := func(name string) {
		n := cat[name]
		{
			n.IsBuilder = true
		}
		cat[name] = n
	}

	addType("A", "", nil)

	addType("MX", "", []Field{
		{Name: "Preference", LegacyName: "MxPreference"},
	})

	addType("SRV", "",
		[]Field{
			{Name: "Priority", Tags: MustParseTags(`json:"priority"`), LegacyName: "SrvPriority"},
			{Name: "Weight", Tags: MustParseTags(`json:"weight"`), LegacyName: "SrvWeight"},
			{Name: "Port", Tags: MustParseTags(`json:"port"`), LegacyName: "SrvPort"},
			{Name: "Target", Tags: MustParseTags(`json:"target" dns:"domain-name"`), LegacyName: "target"},
		},
	)

	addType("CNAME", "", nil)

	addType("CFSINGLEREDIRECT", "CF_SINGLE_REDIRECT",
		[]Field{
			{Name: "SRDisplay", LegacyName: "target"},
		},
	)
	setNoLabel("CFSINGLEREDIRECT")
	setTTL1("CFSINGLEREDIRECT")

	addType("CAA", "",
		[]Field{
			{Name: "Flag", LegacyName: "CaaFlag"},
			{Name: "Tag", LegacyName: "CaaTag"},
			{Name: "Value", Tags: MustParseTags(`dnscontrol:"_,anyascii"`), LegacyName: "target"},
		},
	)
	setIsBuilder("CAA")

	addType("DS", "",
		[]Field{
			{Name: "KeyTag", LegacyName: "DsKeyTag"},
			{Name: "Algorithm", LegacyName: "DsAlgorithm"},
			{Name: "DigestType", LegacyName: "DsDigestType"},
			{Name: "Digest", LegacyName: "DsDigest", Tags: MustParseTags(`dnscontrol:"_,target,alllower"`)},
		},
	)
	addType("DNSKEY", "",
		[]Field{
			{Name: "Flags", LegacyName: "DnskeyFlags"},
			{Name: "Protocol", LegacyName: "DnskeyProtocol"},
			{Name: "Algorithm", LegacyName: "DnskeyAlgorithm"},
			{Name: "PublicKey", LegacyName: "DnskeyPublicKey"},
		},
	)

	//x, _ := json.MarshalIndent(cat, "", "    ")
	//fmt.Printf("DEBUG: Hints: %s\n", x)
	return l, cat
}
