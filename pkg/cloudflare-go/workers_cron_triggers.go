package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

// WorkerCronTriggerResponse represents the response from the Worker cron trigger
// API endpoint.
type WorkerCronTriggerResponse struct {
	Response
	Result WorkerCronTriggerSchedules `json:"result"`
}

// WorkerCronTriggerSchedules contains the schedule of Worker cron triggers.
type WorkerCronTriggerSchedules struct {
	Schedules []WorkerCronTrigger `json:"schedules"`
}

// WorkerCronTrigger holds an individual cron schedule for a worker.
type WorkerCronTrigger struct {
	Cron       string     `json:"cron"`
	CreatedOn  *time.Time `json:"created_on,omitempty"`
	ModifiedOn *time.Time `json:"modified_on,omitempty"`
}

type ListWorkerCronTriggersParams struct {
	ScriptName string
}

type UpdateWorkerCronTriggersParams struct {
	ScriptName string
	Crons      []WorkerCronTrigger
}

// ListWorkerCronTriggers fetches all available cron triggers for a single Worker
// script.
//
// API reference: https://developers.cloudflare.com/api/operations/worker-cron-trigger-get-cron-triggers
func (api *API) ListWorkerCronTriggers(ctx context.Context, rc *ResourceContainer, params ListWorkerCronTriggersParams) ([]WorkerCronTrigger, error) {
	if rc.Level != AccountRouteLevel {
		return []WorkerCronTrigger{}, ErrRequiredAccountLevelResourceContainer
	}

	if rc.Identifier == "" {
		return []WorkerCronTrigger{}, ErrMissingIdentifier
	}

	uri := fmt.Sprintf("/accounts/%s/workers/scripts/%s/schedules", rc.Identifier, params.ScriptName)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []WorkerCronTrigger{}, err
	}

	result := WorkerCronTriggerResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return []WorkerCronTrigger{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result.Result.Schedules, err
}

// UpdateWorkerCronTriggers updates a single schedule for a Worker cron trigger.
//
// API reference: https://developers.cloudflare.com/api/operations/worker-cron-trigger-update-cron-triggers
func (api *API) UpdateWorkerCronTriggers(ctx context.Context, rc *ResourceContainer, params UpdateWorkerCronTriggersParams) ([]WorkerCronTrigger, error) {
	if rc.Level != AccountRouteLevel {
		return []WorkerCronTrigger{}, ErrRequiredAccountLevelResourceContainer
	}

	if rc.Identifier == "" {
		return []WorkerCronTrigger{}, ErrMissingIdentifier
	}

	uri := fmt.Sprintf("/accounts/%s/workers/scripts/%s/schedules", rc.Identifier, params.ScriptName)
	res, err := api.makeRequestContext(ctx, http.MethodPut, uri, params.Crons)
	if err != nil {
		return []WorkerCronTrigger{}, err
	}

	result := WorkerCronTriggerResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return []WorkerCronTrigger{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result.Result.Schedules, err
}
