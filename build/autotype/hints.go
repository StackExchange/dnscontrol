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

	addType("A", "", nil)

	addType("MX", "", nil)

	addType("SRV", "",
		[]Field{
			{Name: "Priority", Tags: `json:"priority"`},
			{Name: "Weight", Tags: `json:"weight"`},
			{Name: "Port", Tags: `json:"port"`},
			{Name: "Target", Tags: `json:"target" dns:"domain-name"`},
		})

	// addType("CFSINGLEREDIRECT", "CF_SINGLE_REDIRECT",
	// 	[]Field{
	// 		{Name: "Code", Tags: `json:"code,omitempty"`},
	// 		{Name: "SRName", Tags: `json:"sr_name,omitempty"`},
	// 		{Name: "SRWhen", Tags: `json:"sr_when,omitempty"`},
	// 		{Name: "SRThen", Tags: `json:"sr_then,omitempty"`},
	// 		{Name: "SRRRulesetID", Tags: `json:"sr_rulesetid,omitempty"`},
	// 		{Name: "SRRRulesetRuleID", Tags: `json:"sr_rulesetruleid,omitempty"`},
	// 		{Name: "SRDisplay", Tags: `json:"sr_display,omitempty"`},
	// 		{Name: "PRWhen", Tags: `dns:"skip" json:"pr_when,omitempty"`},
	// 		{Name: "PRThen", Tags: `dns:"skip" json:"pr_then,omitempty"`},
	// 		{Name: "PRPriority", Tags: `dns:"skip" json:"pr_priority,omitempty"`},
	// 		{Name: "PRDisplay", Tags: `dns:"skip" json:"pr_display,omitempty"`},
	// 	},
	// )

	return l, cat
}
