package spf

type dnsCache map[string]map[string][]string

func (c *dnsCache) dnsGet(label, rtype string) ([]string, bool) {
	v1, ok := (*c)[label]
	if !ok {
		return nil, false
	}
	v2, ok := v1[rtype]
	if !ok {
		return nil, false
	}
	return v2, true
}

func (c *dnsCache) dnsPut(label, rtype string, answers []string) {
	if *c == nil {
		*c = make(dnsCache)
	}
	_, ok := (*c)[label]
	if !ok {
		(*c)[label] = make(map[string][]string)
	}
	(*c)[label][rtype] = answers
}
