package main

// ReadHints returns the "hints" configuration as a data structure.  (Right now
// it returns hardcoded constants.  In the future it should read a YAML file.)
func GetHints() TypeCatalog {

	// Some
	return TypeCatalog{

		"A": RTypeConfig{},

		"MX": RTypeConfig{},

		"SRV": RTypeConfig{
			Fields: []Field{
				{Name: "Priority", Tags: `json:"priority"`},
				{Name: "Weight", Tags: `json:"weight"`},
				{Name: "Port", Tags: `json:"port"`},
				{Name: "Target", Tags: `json:"target"`},
			},
		},

		"CFSINGLEREDIRECT": RTypeConfig{
			Token: "CF_SINGLE_REDIRECT",
			Fields: []Field{
				{Name: "Code", Tags: `json:"code,omitempty"`},
				{Name: "SRName", Tags: `json:"sr_name,omitempty"`},
				{Name: "SRWhen", Tags: `json:"sr_when,omitempty"`},
				{Name: "SRThen", Tags: `json:"sr_then,omitempty"`},
				{Name: "SRRRulesetID", Tags: `json:"sr_rulesetid,omitempty"`},
				{Name: "SRRRulesetRuleID", Tags: `json:"sr_rulesetruleid,omitempty"`},
				{Name: "SRDisplay", Tags: `json:"sr_display,omitempty"`},
				{Name: "PRWhen", Tags: `dns:"skip" json:"pr_when,omitempty"`},
				{Name: "PRThen", Tags: `dns:"skip" json:"pr_then,omitempty"`},
				{Name: "PRPriority", Tags: `dns:"skip" json:"pr_priority,omitempty"`},
				{Name: "PRDisplay", Tags: `dns:"skip" json:"pr_display,omitempty"`},
			},
		},
	}
}
