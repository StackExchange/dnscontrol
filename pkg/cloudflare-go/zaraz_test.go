package cloudflare

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var trueValue bool = true
var falseValue bool = false

var expectedConfig ZarazConfig = ZarazConfig{
	DebugKey:      "cheese",
	ZarazVersion:  44,
	DataLayer:     &trueValue,
	Dlp:           []any{},
	HistoryChange: &trueValue,
	Settings: ZarazConfigSettings{
		AutoInjectScript: &trueValue,
	},
	Tools: map[string]ZarazTool{
		"PBQr": {
			BlockingTriggers: []string{},
			Enabled:          &trueValue,
			DefaultFields:    map[string]any{},
			Name:             "Custom HTML",
			Actions: map[string]ZarazAction{
				"7ccae28d-5e00-4f0b-a491-519ecde998c8": {
					ActionType:       "event",
					BlockingTriggers: []string{},
					Data: map[string]any{
						"__zaraz_setting_name": "pageview1",
						"htmlCode":             "<script>console.log('pageview')</script>",
					},
					FiringTriggers: []string{"Pageview"},
				},
			},
			Type:           ZarazToolComponent,
			DefaultPurpose: "rJJC",
			Component:      "html",
			Permissions:    []string{"execute_unsafe_scripts"},
			Settings:       map[string]any{},
		},
	},
	Triggers: map[string]ZarazTrigger{
		"Pageview": {
			Name:         "Pageview",
			Description:  "All page loads",
			LoadRules:    []ZarazTriggerRule{{Match: "{{ client.__zarazTrack }}", Op: "EQUALS", Value: "Pageview"}},
			ExcludeRules: []ZarazTriggerRule{},
			ClientRules:  []any{},
			System:       ZarazPageload,
		},
		"TFOl": {
			Name:         "test",
			Description:  "",
			LoadRules:    []ZarazTriggerRule{{Id: "Kqsc", Match: "test", Op: "CONTAINS", Value: "test"}, {Id: "EDnV", Action: ZarazClickListener, Settings: ZarazRuleSettings{Selector: "test", Type: ZarazCSS}}},
			ExcludeRules: []ZarazTriggerRule{},
		},
	},
	Variables: map[string]ZarazVariable{
		"jwIx": {
			Name:  "test",
			Type:  ZarazVarString,
			Value: "sss",
		},
		"pAuL": {
			Name: "test-worker-var",
			Type: ZarazVarWorker,
			Value: map[string]interface{}{
				"escapedWorkerName": "worker-var-example",
				"mutableId":         "m.zpt3q__WyW-61WM2qwgGoBl4Nxg-sfBsaMhu9NayjwU",
				"workerTag":         "68aba570db9d4ec5b159624e2f7ad8bf",
			},
		},
	},
	Consent: ZarazConsent{
		Enabled: &trueValue,
		ButtonTextTranslations: ZarazButtonTextTranslations{
			AcceptAll:        map[string]string{"en": "Accept ALL"},
			ConfirmMyChoices: map[string]string{"en": "YES!"},
			RejectAll:        map[string]string{"en": "Reject ALL"},
		},
		CompanyEmail:                          "email@example.com",
		ConsentModalIntroHTMLWithTranslations: map[string]string{"en": "Lorem ipsum dolar set Amet?"},
		CookieName:                            "zaraz-consent",
		CustomCSS:                             ".test {\n    color: red;\n}",
		CustomIntroDisclaimerDismissed:        &trueValue,
		DefaultLanguage:                       "en",
		HideModal:                             &falseValue,
		PurposesWithTranslations: map[string]ZarazPurposeWithTranslations{
			"rJJC": {
				Description: map[string]string{"en": "Blah blah"},
				Name:        map[string]string{"en": "Analytics"},
				Order:       0,
			},
		},
	},
}

func TestGetZarazConfig(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")

		fmt.Fprint(w, `{
			"errors": [],
			"messages": [],
			"success": true,
			"result": {
				"consent": {
				  "buttonTextTranslations": {
					"accept_all": {
					  "en": "Accept ALL"
					},
					"confirm_my_choices": {
					  "en": "YES!"
					},
					"reject_all": {
					  "en": "Reject ALL"
					}
				  },
				  "companyEmail": "email@example.com",
				  "consentModalIntroHTMLWithTranslations": {
					"en": "Lorem ipsum dolar set Amet?"
				  },
				  "cookieName": "zaraz-consent",
				  "customCSS": ".test {\n    color: red;\n}",
				  "customIntroDisclaimerDismissed": true,
				  "defaultLanguage": "en",
				  "enabled": true,
				  "hideModal": false,
				  "purposesWithTranslations": {
					"rJJC": {
					  "description": {
						"en": "Blah blah"
					  },
					  "name": {
						"en": "Analytics"
					  },
					  "order": 0
					}
				  }
				},
				"dataLayer": true,
				"debugKey": "cheese",
				"dlp": [],
				"historyChange": true,
				"invalidKey": "cheese",
				"settings": {
				  "autoInjectScript": true
				},
				"tools": {
				  "PBQr": {
					"blockingTriggers": [],
					"component": "html",
					"defaultFields": {},
					"defaultPurpose": "rJJC",
					"enabled": true,
					"mode": {
					  "cloud": false,
					  "ignoreSPA": true,
					  "light": false,
					  "sample": false,
					  "segment": {
						"end": 100,
						"start": 0
					  },
					  "trigger": "pageload"
					},
					"name": "Custom HTML",
					"actions": {
					  "7ccae28d-5e00-4f0b-a491-519ecde998c8": {
							"actionType": "event",
							"blockingTriggers": [],
							"data": {
								"__zaraz_setting_name": "pageview1",
								"htmlCode": "<script>console.log('pageview')</script>"
							},
							"firingTriggers": [
								"Pageview"
							]
					  }
					},
					"permissions": [
					  "execute_unsafe_scripts"
					],
					"settings": {},
					"type": "component"
				  }
				},
				"triggers": {
				  "Pageview": {
					"clientRules": [],
					"description": "All page loads",
					"excludeRules": [],
					"loadRules": [
					  {
						"match": "{{ client.__zarazTrack }}",
						"op": "EQUALS",
						"value": "Pageview"
					  }
					],
					"name": "Pageview",
					"system": "pageload"
				  },
				  "TFOl": {
					"description": "",
					"excludeRules": [],
					"loadRules": [
					  {
						"id": "Kqsc",
						"match": "test",
						"op": "CONTAINS",
						"value": "test"
					  },
					  {
						"action": "clickListener",
						"id": "EDnV",
						"settings": {
						  "selector": "test",
						  "type": "css",
						  "waitForTags": 0
						}
					  }
					],
					"name": "test"
				  }
				},
				"variables": {
				  "jwIx": {
					"name": "test",
					"type": "string",
					"value": "sss"
				  },
				  "pAuL": {
					"name": "test-worker-var",
					"type": "worker",
					"value": {
					  "escapedWorkerName": "worker-var-example",
					  "mutableId": "m.zpt3q__WyW-61WM2qwgGoBl4Nxg-sfBsaMhu9NayjwU",
					  "workerTag": "68aba570db9d4ec5b159624e2f7ad8bf"
					}
				  }
				},
				"zarazVersion": 44
			  }
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/settings/zaraz/v2/config", handler)

	actual, err := client.GetZarazConfig(context.Background(), ZoneIdentifier(testZoneID))
	require.NoError(t, err)

	assert.Equal(t, expectedConfig, actual.Result)
}

func TestUpdateZarazConfig(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"errors": [],
			"messages": [],
			"success": true,
			"result": {
				"consent": {
				  "buttonTextTranslations": {
					"accept_all": {
					  "en": "Accept ALL"
					},
					"confirm_my_choices": {
					  "en": "YES!"
					},
					"reject_all": {
					  "en": "Reject ALL"
					}
				  },
				  "companyEmail": "email@example.com",
				  "consentModalIntroHTMLWithTranslations": {
					"en": "Lorem ipsum dolar set Amet?"
				  },
				  "cookieName": "zaraz-consent",
				  "customCSS": ".test {\n    color: red;\n}",
				  "customIntroDisclaimerDismissed": true,
				  "defaultLanguage": "en",
				  "enabled": true,
				  "hideModal": false,
				  "purposesWithTranslations": {
					"rJJC": {
					  "description": {
						"en": "Blah blah"
					  },
					  "name": {
						"en": "Analytics"
					  },
					  "order": 0
					}
				  }
				},
				"dataLayer": true,
				"debugKey": "butter",
				"dlp": [],
				"historyChange": true,
				"invalidKey": "cheese",
				"settings": {
				  "autoInjectScript": true
				},
				"tools": {
				  "PBQr": {
					"blockingTriggers": [],
					"component": "html",
					"defaultFields": {},
					"defaultPurpose": "rJJC",
					"enabled": true,
					"mode": {
					  "cloud": false,
					  "ignoreSPA": true,
					  "light": false,
					  "sample": false,
					  "segment": {
						"end": 100,
						"start": 0
					  },
					  "trigger": "pageload"
					},
					"name": "Custom HTML",
					"actions": {
					  "7ccae28d-5e00-4f0b-a491-519ecde998c8": {
							"actionType": "event",
							"blockingTriggers": [],
							"data": {
								"__zaraz_setting_name": "pageview1",
								"htmlCode": "<script>console.log('pageview')</script>"
							},
							"firingTriggers": [
								"Pageview"
							]
					  }
					},
					"permissions": [
					  "execute_unsafe_scripts"
					],
					"settings": {},
					"type": "component"
				  }
				},
				"triggers": {
				  "Pageview": {
					"clientRules": [],
					"description": "All page loads",
					"excludeRules": [],
					"loadRules": [
					  {
						"match": "{{ client.__zarazTrack }}",
						"op": "EQUALS",
						"value": "Pageview"
					  }
					],
					"name": "Pageview",
					"system": "pageload"
				  },
				  "TFOl": {
					"description": "",
					"excludeRules": [],
					"loadRules": [
					  {
						"id": "Kqsc",
						"match": "test",
						"op": "CONTAINS",
						"value": "test"
					  },
					  {
						"action": "clickListener",
						"id": "EDnV",
						"settings": {
						  "selector": "test",
						  "type": "css",
						  "waitForTags": 0
						}
					  }
					],
					"name": "test"
				  }
				},
				"variables": {
				  "jwIx": {
					"name": "test",
					"type": "string",
					"value": "sss"
				  },
				  "pAuL": {
					"name": "test-worker-var",
					"type": "worker",
					"value": {
					  "escapedWorkerName": "worker-var-example",
					  "mutableId": "m.zpt3q__WyW-61WM2qwgGoBl4Nxg-sfBsaMhu9NayjwU",
					  "workerTag": "68aba570db9d4ec5b159624e2f7ad8bf"
					}
				  }
				},
				"zarazVersion": 44
			  }
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/settings/zaraz/v2/config", handler)
	payload := UpdateZarazConfigParams{
		DebugKey:      "cheese",
		ZarazVersion:  44,
		DataLayer:     &trueValue,
		Dlp:           []any{},
		HistoryChange: &trueValue,
		Settings: ZarazConfigSettings{
			AutoInjectScript: &trueValue,
		},
		Tools: map[string]ZarazTool{
			"PBQr": {
				BlockingTriggers: []string{},
				Enabled:          &trueValue,
				DefaultFields:    map[string]any{},
				Name:             "Custom HTML",
				Actions: map[string]ZarazAction{
					"7ccae28d-5e00-4f0b-a491-519ecde998c8": {
						ActionType:       "event",
						BlockingTriggers: []string{},
						Data: map[string]any{
							"__zaraz_setting_name": "pageview1",
							"htmlCode":             "<script>console.log('pageview')</script>",
						},
						FiringTriggers: []string{"Pageview"},
					},
				},
				Type:           ZarazToolComponent,
				DefaultPurpose: "rJJC",
				Component:      "html",
				Permissions:    []string{"execute_unsafe_scripts"},
				Settings:       map[string]any{},
			},
		},
		Triggers: map[string]ZarazTrigger{
			"Pageview": {
				Name:         "Pageview",
				Description:  "All page loads",
				LoadRules:    []ZarazTriggerRule{{Match: "{{ client.__zarazTrack }}", Op: "EQUALS", Value: "Pageview"}},
				ExcludeRules: []ZarazTriggerRule{},
				ClientRules:  []any{},
				System:       ZarazPageload,
			},
			"TFOl": {
				Name:         "test",
				Description:  "",
				LoadRules:    []ZarazTriggerRule{{Id: "Kqsc", Match: "test", Op: "CONTAINS", Value: "test"}, {Id: "EDnV", Action: ZarazClickListener, Settings: ZarazRuleSettings{Selector: "test", Type: ZarazCSS}}},
				ExcludeRules: []ZarazTriggerRule{},
			},
		},
		Variables: map[string]ZarazVariable{
			"jwIx": {
				Name:  "test",
				Type:  ZarazVarString,
				Value: "sss",
			},
			"pAuL": {
				Name: "test-worker-var",
				Type: ZarazVarWorker,
				Value: map[string]interface{}{
					"escapedWorkerName": "worker-var-example",
					"mutableId":         "m.zpt3q__WyW-61WM2qwgGoBl4Nxg-sfBsaMhu9NayjwU",
					"workerTag":         "68aba570db9d4ec5b159624e2f7ad8bf",
				},
			},
		},
		Consent: ZarazConsent{
			Enabled: &trueValue,
			ButtonTextTranslations: ZarazButtonTextTranslations{
				AcceptAll:        map[string]string{"en": "Accept ALL"},
				ConfirmMyChoices: map[string]string{"en": "YES!"},
				RejectAll:        map[string]string{"en": "Reject ALL"},
			},
			CompanyEmail:                          "email@example.com",
			ConsentModalIntroHTMLWithTranslations: map[string]string{"en": "Lorem ipsum dolar set Amet?"},
			CookieName:                            "zaraz-consent",
			CustomCSS:                             ".test {\n    color: red;\n}",
			CustomIntroDisclaimerDismissed:        &trueValue,
			DefaultLanguage:                       "en",
			HideModal:                             &falseValue,
			PurposesWithTranslations: map[string]ZarazPurposeWithTranslations{
				"rJJC": {
					Description: map[string]string{"en": "Blah blah"},
					Name:        map[string]string{"en": "Analytics"},
					Order:       0,
				},
			},
		},
	}
	payload.DebugKey = "butter" // Updating config
	modifiedConfig := expectedConfig
	modifiedConfig.DebugKey = "butter" // Updating config
	expected := ZarazConfigResponse{
		Result: modifiedConfig,
	}

	actual, err := client.UpdateZarazConfig(context.Background(), ZoneIdentifier(testZoneID), payload)
	require.NoError(t, err)

	assert.Equal(t, expected.Result, actual.Result)
}

func TestUpdateDeprecatedZarazConfig(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"errors": [],
			"messages": [],
			"success": true,
			"result": {
				"consent": {
				  "buttonTextTranslations": {
					"accept_all": {
					  "en": "Accept ALL"
					},
					"confirm_my_choices": {
					  "en": "YES!"
					},
					"reject_all": {
					  "en": "Reject ALL"
					}
				  },
				  "companyEmail": "email@example.com",
				  "consentModalIntroHTMLWithTranslations": {
					"en": "Lorem ipsum dolar set Amet?"
				  },
				  "cookieName": "zaraz-consent",
				  "customCSS": ".test {\n    color: red;\n}",
				  "customIntroDisclaimerDismissed": true,
				  "defaultLanguage": "en",
				  "enabled": true,
				  "hideModal": false,
				  "purposesWithTranslations": {
					"rJJC": {
					  "description": {
						"en": "Blah blah"
					  },
					  "name": {
						"en": "Analytics"
					  },
					  "order": 0
					}
				  }
				},
				"dataLayer": true,
				"debugKey": "butter",
				"dlp": [],
				"historyChange": true,
				"invalidKey": "cheese",
				"settings": {
				  "autoInjectScript": true
				},
				"tools": {
				  "PBQr": {
					"blockingTriggers": [],
					"component": "html",
					"defaultFields": {},
					"defaultPurpose": "rJJC",
					"enabled": true,
					"mode": {
					  "cloud": false,
					  "ignoreSPA": true,
					  "light": false,
					  "sample": false,
					  "segment": {
						"end": 100,
						"start": 0
					  },
					  "trigger": "pageload"
					},
					"name": "Custom HTML",
					"actions": {
					  "7ccae28d-5e00-4f0b-a491-519ecde998c8": {
							"actionType": "event",
							"blockingTriggers": [],
							"data": {
								"__zaraz_setting_name": "pageview1",
								"htmlCode": "<script>console.log('pageview')</script>"
							},
							"firingTriggers": [
								"Pageview"
							]
					  }
					},
					"permissions": [
					  "execute_unsafe_scripts"
					],
					"settings": {},
					"type": "component"
				  }
				},
				"triggers": {
				  "Pageview": {
					"clientRules": [],
					"description": "All page loads",
					"excludeRules": [],
					"loadRules": [
					  {
						"match": "{{ client.__zarazTrack }}",
						"op": "EQUALS",
						"value": "Pageview"
					  }
					],
					"name": "Pageview",
					"system": "pageload"
				  },
				  "TFOl": {
					"description": "",
					"excludeRules": [],
					"loadRules": [
					  {
						"id": "Kqsc",
						"match": "test",
						"op": "CONTAINS",
						"value": "test"
					  },
					  {
						"action": "clickListener",
						"id": "EDnV",
						"settings": {
						  "selector": "test",
						  "type": "css",
						  "waitForTags": 0
						}
					  }
					],
					"name": "test"
				  }
				},
				"variables": {
				  "jwIx": {
					"name": "test",
					"type": "string",
					"value": "sss"
				  },
				  "pAuL": {
					"name": "test-worker-var",
					"type": "worker",
					"value": {
					  "escapedWorkerName": "worker-var-example",
					  "mutableId": "m.zpt3q__WyW-61WM2qwgGoBl4Nxg-sfBsaMhu9NayjwU",
					  "workerTag": "68aba570db9d4ec5b159624e2f7ad8bf"
					}
				  }
				},
				"zarazVersion": 44
			  }
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/settings/zaraz/v2/config", handler)
	payload := UpdateZarazConfigParams{
		DebugKey:      "cheese",
		ZarazVersion:  44,
		DataLayer:     &trueValue,
		Dlp:           []any{},
		HistoryChange: &trueValue,
		Settings: ZarazConfigSettings{
			AutoInjectScript: &trueValue,
		},
		Tools: map[string]ZarazTool{
			"PBQr": {
				BlockingTriggers: []string{},
				Enabled:          &trueValue,
				DefaultFields:    map[string]any{},
				Name:             "Custom HTML",
				Actions: map[string]ZarazAction{
					"7ccae28d-5e00-4f0b-a491-519ecde998c8": {
						ActionType:       "event",
						BlockingTriggers: []string{},
						Data: map[string]any{
							"__zaraz_setting_name": "pageview1",
							"htmlCode":             "<script>console.log('pageview')</script>",
						},
						FiringTriggers: []string{"Pageview"},
					},
				},
				Type:           ZarazToolComponent,
				DefaultPurpose: "rJJC",
				Component:      "html",
				Permissions:    []string{"execute_unsafe_scripts"},
				Settings:       map[string]any{},
			},
		},
		Triggers: map[string]ZarazTrigger{
			"Pageview": {
				Name:         "Pageview",
				Description:  "All page loads",
				LoadRules:    []ZarazTriggerRule{{Match: "{{ client.__zarazTrack }}", Op: "EQUALS", Value: "Pageview"}},
				ExcludeRules: []ZarazTriggerRule{},
				ClientRules:  []any{},
				System:       ZarazPageload,
			},
			"TFOl": {
				Name:         "test",
				Description:  "",
				LoadRules:    []ZarazTriggerRule{{Id: "Kqsc", Match: "test", Op: "CONTAINS", Value: "test"}, {Id: "EDnV", Action: ZarazClickListener, Settings: ZarazRuleSettings{Selector: "test", Type: ZarazCSS}}},
				ExcludeRules: []ZarazTriggerRule{},
			},
		},
		Variables: map[string]ZarazVariable{
			"jwIx": {
				Name:  "test",
				Type:  ZarazVarString,
				Value: "sss",
			},
			"pAuL": {
				Name: "test-worker-var",
				Type: ZarazVarWorker,
				Value: map[string]interface{}{
					"escapedWorkerName": "worker-var-example",
					"mutableId":         "m.zpt3q__WyW-61WM2qwgGoBl4Nxg-sfBsaMhu9NayjwU",
					"workerTag":         "68aba570db9d4ec5b159624e2f7ad8bf",
				},
			},
		},
		Consent: ZarazConsent{
			Enabled: &trueValue,
			ButtonTextTranslations: ZarazButtonTextTranslations{
				AcceptAll:        map[string]string{"en": "Accept ALL"},
				ConfirmMyChoices: map[string]string{"en": "YES!"},
				RejectAll:        map[string]string{"en": "Reject ALL"},
			},
			CompanyEmail:                          "email@example.com",
			ConsentModalIntroHTMLWithTranslations: map[string]string{"en": "Lorem ipsum dolar set Amet?"},
			CookieName:                            "zaraz-consent",
			CustomCSS:                             ".test {\n    color: red;\n}",
			CustomIntroDisclaimerDismissed:        &trueValue,
			DefaultLanguage:                       "en",
			HideModal:                             &falseValue,
			PurposesWithTranslations: map[string]ZarazPurposeWithTranslations{
				"rJJC": {
					Description: map[string]string{"en": "Blah blah"},
					Name:        map[string]string{"en": "Analytics"},
					Order:       0,
				},
			},
		},
	}
	payload.DebugKey = "butter" // Updating config
	modifiedConfig := expectedConfig
	modifiedConfig.DebugKey = "butter" // Updating config
	expected := ZarazConfigResponse{
		Result: modifiedConfig,
	}

	actual, err := client.UpdateZarazConfig(context.Background(), ZoneIdentifier(testZoneID), payload)
	require.NoError(t, err)

	assert.Equal(t, expected.Result, actual.Result)
}

func TestGetZarazWorkflow(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"errors": [],
			"messages": [],
			"success": true,
			"result": "realtime"
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/settings/zaraz/v2/workflow", handler)
	want := ZarazWorkflowResponse{
		Result: "realtime",
		Response: Response{
			Success:  true,
			Messages: []ResponseInfo{},
			Errors:   []ResponseInfo{},
		},
	}

	actual, err := client.GetZarazWorkflow(context.Background(), ZoneIdentifier(testZoneID))

	require.NoError(t, err)

	assert.Equal(t, want, actual)
}

func TestUpdateZarazWorkflow(t *testing.T) {
	setup()
	defer teardown()

	payload := UpdateZarazWorkflowParams{
		Workflow: "realtime",
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		body, _ := io.ReadAll(r.Body)
		bodyString := string(body)
		assert.Equal(t, fmt.Sprintf("\"%s\"", payload.Workflow), bodyString)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"errors": [],
			"messages": [],
			"success": true,
			"result": "realtime"
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/settings/zaraz/v2/workflow", handler)
	want := ZarazWorkflowResponse{
		Result: "realtime",
		Response: Response{
			Success:  true,
			Messages: []ResponseInfo{},
			Errors:   []ResponseInfo{},
		},
	}

	actual, err := client.UpdateZarazWorkflow(context.Background(), ZoneIdentifier(testZoneID), payload)

	require.NoError(t, err)

	assert.Equal(t, want, actual)
}

func TestPublishZarazConfig(t *testing.T) {
	setup()
	defer teardown()

	payload := PublishZarazConfigParams{
		Description: "test description",
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		body, _ := io.ReadAll(r.Body)
		bodyString := string(body)
		assert.Equal(t, fmt.Sprintf("\"%s\"", payload.Description), bodyString)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"errors": [],
			"messages": [],
			"success": true,
			"result": "Config has been published successfully"
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/settings/zaraz/v2/publish", handler)
	want := ZarazPublishResponse{
		Result: "Config has been published successfully",
		Response: Response{
			Success:  true,
			Messages: []ResponseInfo{},
			Errors:   []ResponseInfo{},
		},
	}

	actual, err := client.PublishZarazConfig(context.Background(), ZoneIdentifier(testZoneID), payload)

	require.NoError(t, err)

	assert.Equal(t, want, actual)
}

func TestGetZarazConfigHistoryList(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": [
			  {
				"createdAt": "2023-01-01T05:20:00Z",
				"description": "test 1",
				"id": 1005736,
				"updatedAt": "2023-01-01T05:20:00Z",
				"userId": "9ceddf6f117afe04c64716c83468d3a4"
			  },
			  {
				"createdAt": "2023-01-01T05:20:00Z",
				"description": "test 2",
				"id": 1005735,
				"updatedAt": "2023-01-01T05:20:00Z",
				"userId": "9ceddf6f117afe04c64716c83468d3a4"
			  }
			],
			"success": true,
			"errors": [],
			"messages": [],
			"result_info": {
			  "page": 1,
			  "per_page": 2,
			  "count": 2,
			  "total_count": 8
			}
		  }`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/settings/zaraz/v2/history", handler)
	createdAt, _ := time.Parse(time.RFC3339, "2023-01-01T05:20:00Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2023-01-01T05:20:00Z")
	want := []ZarazHistoryRecord{
		{
			CreatedAt:   &createdAt,
			Description: "test 1",
			ID:          1005736,
			UpdatedAt:   &updatedAt,
			UserID:      "9ceddf6f117afe04c64716c83468d3a4",
		},
		{
			CreatedAt:   &createdAt,
			Description: "test 2",
			ID:          1005735,
			UpdatedAt:   &updatedAt,
			UserID:      "9ceddf6f117afe04c64716c83468d3a4",
		},
	}

	actual, _, err := client.ListZarazConfigHistory(context.Background(), ZoneIdentifier(testZoneID), ListZarazConfigHistoryParams{
		ResultInfo: ResultInfo{
			PerPage: 2,
		},
	})

	require.NoError(t, err)

	assert.Equal(t, want, actual)
}

func TestGetDefaultZarazConfig(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "text/plain")
		fmt.Fprint(w, `{
			"someTestKeyThatRepsTheConfig": "test"
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/settings/zaraz/v2/default", handler)

	_, err := client.GetDefaultZarazConfig(context.Background(), ZoneIdentifier(testZoneID))
	require.NoError(t, err)
}

func TestExportZarazConfig(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "text/plain")
		fmt.Fprint(w, `{
			"someTestKeyThatRepsTheConfig": "test"
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/settings/zaraz/v2/export", handler)

	err := client.ExportZarazConfig(context.Background(), ZoneIdentifier(testZoneID))
	require.NoError(t, err)
}
