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

	// 1
	addType("A", "", nil)

	// 2
	addType("NS", "", nil)

	// 5
	addType("CNAME", "", nil)

	// 6
	//addType("SOA", "", nil)

	// 12
	addType("PTR", "", nil)

	// 15
	addType("MX", "", []Field{
		{Name: "Preference", LegacyName: "MxPreference"},
	})

	// 16
	//addType("TXT", "", []Field{
	//	{Name: "Text", LegacyName: "Text", Tags: MustParseTags(`dns:"txtsegments"`)},
	//})

	// 28
	addType("AAAA", "", []Field{
		{Name: "AAAA", Tags: MustParseTags(`dns:"aaaa"`)},
	})

	// 29
	//addType("LOC", "", nil)

	// 33
	addType("SRV", "", []Field{
		{Name: "Priority", Tags: MustParseTags(`json:"priority"`), LegacyName: "SrvPriority"},
		{Name: "Weight", Tags: MustParseTags(`json:"weight"`), LegacyName: "SrvWeight"},
		{Name: "Port", Tags: MustParseTags(`json:"port"`), LegacyName: "SrvPort"},
		{Name: "Target", Tags: MustParseTags(`json:"target" dns:"domain-name"`), LegacyName: "target"},
	},
	)

	// 35
	addType("NAPTR", "", []Field{
		{Name: "Order", LegacyName: "NaptrOrder"},
		{Name: "Preference", LegacyName: "NaptrPreference"},
		{Name: "Flags", LegacyName: "NaptrFlags", Tags: MustParseTags(`dnscontrol:"_,anyascii"`)},
		{Name: "Service", LegacyName: "NaptrService", Tags: MustParseTags(`dnscontrol:"_,anyascii"`)},
		{Name: "Regexp", LegacyName: "NaptrRegexp", Tags: MustParseTags(`dnscontrol:"_,anyascii"`)},
		{Name: "Replacement", LegacyName: "target", Tags: MustParseTags(`dnscontrol:"_,empty_becomes_dot"`)},
	},
	)

	// 39
	//addType("DNAME", "", nil)

	// 43
	addType("DS", "", []Field{
		{Name: "KeyTag", LegacyName: "DsKeyTag"},
		{Name: "Algorithm", LegacyName: "DsAlgorithm"},
		{Name: "DigestType", LegacyName: "DsDigestType"},
		{Name: "Digest", LegacyName: "DsDigest", Tags: MustParseTags(`dnscontrol:"_,target,alllower"`)},
	},
	)

	// 44
	//addType("SSHFP", "", nil)

	// 48
	addType("DNSKEY", "", []Field{
		{Name: "Flags", LegacyName: "DnskeyFlags"},
		{Name: "Protocol", LegacyName: "DnskeyProtocol"},
		{Name: "Algorithm", LegacyName: "DnskeyAlgorithm"},
		{Name: "PublicKey", LegacyName: "DnskeyPublicKey"},
	},
	)

	// 49
	//addType("DHCID", "", nil)

	// 52
	//addType("TLSA", "", nil)

	// 61
	//addType("OPENPGPKEY", "", nil)

	// 64
	//addType("SVCB", "", nil)

	// 65
	//addType("HTTPS", "", nil)

	// 257
	addType("CAA", "", []Field{
		{Name: "Flag", LegacyName: "CaaFlag"},
		{Name: "Tag", LegacyName: "CaaTag"},
		{Name: "Value", Tags: MustParseTags(`dnscontrol:"_,anyascii"`), LegacyName: "target"},
	},
	)
	setIsBuilder("CAA")

	addType("CFSINGLEREDIRECT", "CF_SINGLE_REDIRECT",
		[]Field{
			{Name: "SRDisplay", LegacyName: "target"},
		})
	setNoLabel("CFSINGLEREDIRECT")
	setTTL1("CFSINGLEREDIRECT")

	//x, _ := json.MarshalIndent(cat, "", "    ")
	//fmt.Printf("DEBUG: Hints: %s\n", x)
	return l, cat
}
