package cloudflare

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	expectedBehaviors = Behaviors{
		Behaviors: map[string]Behavior{
			"high_dlp": {
				Name:        "High Number of DLP Policies Triggered",
				Description: "User has triggered an active DLP profile in a Gateway policy fifteen times or more within one minute.",
				RiskLevel:   Low,
				Enabled:     BoolPtr(true),
			},
			"imp_travel": {
				Name:        "Impossible Travel",
				Description: "A user had a successful Access application log in from two locations that they could not have traveled to in that period of time.",
				RiskLevel:   High,
				Enabled:     BoolPtr(false),
			},
		},
	}

	updateBehaviors = Behaviors{
		Behaviors: map[string]Behavior{
			"high_dlp": {
				RiskLevel: Low,
				Enabled:   BoolPtr(true),
			},
			"imp_travel": {
				RiskLevel: High,
				Enabled:   BoolPtr(false),
			},
		},
	}
)

func TestBehaviors(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
				"result": {
					"behaviors": {
						"high_dlp": {
							"name": "High Number of DLP Policies Triggered",
							"description": "User has triggered an active DLP profile in a Gateway policy fifteen times or more within one minute.",
							"risk_level": "low",
							"enabled":true
						},
						"imp_travel": {
							"name": "Impossible Travel",
							"description": "A user had a successful Access application log in from two locations that they could not have traveled to in that period of time.",
							"risk_level": "high",
							"enabled": false
						}
					}
				},
				"success": true,
				"errors": [],
				"messages": []
			}`)
	}

	mux.HandleFunc("/accounts/01a7362d577a6c3019a474fd6f485823/zt_risk_scoring/behaviors", handler)
	want := expectedBehaviors

	actual, err := client.Behaviors(context.Background(), "01a7362d577a6c3019a474fd6f485823")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateBehaviors(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if assert.NoError(t, err) {
			assert.JSONEq(t, `{
					"behaviors": {
						"high_dlp": {
							"risk_level": "low",
							"enabled":true
						},
						"imp_travel": {
							"risk_level": "high",
							"enabled": false
						}
					}
				}`, string(b), "JSON payload not equal")
		}

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
				"result": {
					"behaviors": {
						"high_dlp": {
							"name": "High Number of DLP Policies Triggered",
							"description": "User has triggered an active DLP profile in a Gateway policy fifteen times or more within one minute.",
							"risk_level": "low",
							"enabled":true
						},
						"imp_travel": {
							"name": "Impossible Travel",
							"description": "A user had a successful Access application log in from two locations that they could not have traveled to in that period of time.",
							"risk_level": "high",
							"enabled": false
						}
					}
				},
				"success": true,
				"errors": [],
				"messages": []
			}`)
	}

	mux.HandleFunc("/accounts/01a7362d577a6c3019a474fd6f485823/zt_risk_scoring/behaviors", handler)

	want := expectedBehaviors
	actual, err := client.UpdateBehaviors(context.Background(), "01a7362d577a6c3019a474fd6f485823", updateBehaviors)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestRiskLevelFromString(t *testing.T) {
	got, _ := RiskLevelFromString("high")
	want := High

	if *got != want {
		t.Errorf("got %#v, wanted %#v", *got, want)
	}
}

func TestStringFromRiskLevel(t *testing.T) {
	got := fmt.Sprint(High)
	want := "high"

	if got != want {
		t.Errorf("got %#v, wanted %#v", got, want)
	}
}
