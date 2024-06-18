package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

// LogpushJob describes a Logpush job.
type LogpushJob struct {
	ID                       int                   `json:"id,omitempty"`
	Dataset                  string                `json:"dataset"`
	Enabled                  bool                  `json:"enabled"`
	Kind                     string                `json:"kind,omitempty"`
	Name                     string                `json:"name"`
	LogpullOptions           string                `json:"logpull_options,omitempty"`
	OutputOptions            *LogpushOutputOptions `json:"output_options,omitempty"`
	DestinationConf          string                `json:"destination_conf"`
	OwnershipChallenge       string                `json:"ownership_challenge,omitempty"`
	LastComplete             *time.Time            `json:"last_complete,omitempty"`
	LastError                *time.Time            `json:"last_error,omitempty"`
	ErrorMessage             string                `json:"error_message,omitempty"`
	Frequency                string                `json:"frequency,omitempty"`
	Filter                   *LogpushJobFilters    `json:"filter,omitempty"`
	MaxUploadBytes           int                   `json:"max_upload_bytes,omitempty"`
	MaxUploadRecords         int                   `json:"max_upload_records,omitempty"`
	MaxUploadIntervalSeconds int                   `json:"max_upload_interval_seconds,omitempty"`
}

type LogpushJobFilters struct {
	Where LogpushJobFilter `json:"where"`
}

type Operator string

const (
	Equal              Operator = "eq"
	NotEqual           Operator = "!eq"
	LessThan           Operator = "lt"
	LessThanOrEqual    Operator = "leq"
	GreaterThan        Operator = "gt"
	GreaterThanOrEqual Operator = "geq"
	StartsWith         Operator = "startsWith"
	EndsWith           Operator = "endsWith"
	NotStartsWith      Operator = "!startsWith"
	NotEndsWith        Operator = "!endsWith"
	Contains           Operator = "contains"
	NotContains        Operator = "!contains"
	ValueIsIn          Operator = "in"
	ValueIsNotIn       Operator = "!in"
)

type LogpushJobFilter struct {
	// either this
	And []LogpushJobFilter `json:"and,omitempty"`
	Or  []LogpushJobFilter `json:"or,omitempty"`
	// or this
	Key      string      `json:"key,omitempty"`
	Operator Operator    `json:"operator,omitempty"`
	Value    interface{} `json:"value,omitempty"`
}

type LogpushOutputOptions struct {
	FieldNames      []string `json:"field_names"`
	OutputType      string   `json:"output_type,omitempty"`
	BatchPrefix     string   `json:"batch_prefix,omitempty"`
	BatchSuffix     string   `json:"batch_suffix,omitempty"`
	RecordPrefix    string   `json:"record_prefix,omitempty"`
	RecordSuffix    string   `json:"record_suffix,omitempty"`
	RecordTemplate  string   `json:"record_template,omitempty"`
	RecordDelimiter string   `json:"record_delimiter,omitempty"`
	FieldDelimiter  string   `json:"field_delimiter,omitempty"`
	TimestampFormat string   `json:"timestamp_format,omitempty"`
	SampleRate      float64  `json:"sample_rate,omitempty"`
	CVE202144228    *bool    `json:"CVE-2021-44228,omitempty"`
}

// LogpushJobsResponse is the API response, containing an array of Logpush Jobs.
type LogpushJobsResponse struct {
	Response
	Result []LogpushJob `json:"result"`
}

// LogpushJobDetailsResponse is the API response, containing a single Logpush Job.
type LogpushJobDetailsResponse struct {
	Response
	Result LogpushJob `json:"result"`
}

// LogpushFieldsResponse is the API response for a datasets fields.
type LogpushFieldsResponse struct {
	Response
	Result LogpushFields `json:"result"`
}

// LogpushFields is a map of available Logpush field names & descriptions.
type LogpushFields map[string]string

// LogpushGetOwnershipChallenge describes a ownership validation.
type LogpushGetOwnershipChallenge struct {
	Filename string `json:"filename"`
	Valid    bool   `json:"valid"`
	Message  string `json:"message"`
}

// LogpushGetOwnershipChallengeResponse is the API response, containing a ownership challenge.
type LogpushGetOwnershipChallengeResponse struct {
	Response
	Result LogpushGetOwnershipChallenge `json:"result"`
}

// LogpushGetOwnershipChallengeRequest is the API request for get ownership challenge.
type LogpushGetOwnershipChallengeRequest struct {
	DestinationConf string `json:"destination_conf"`
}

// LogpushOwnershipChallengeValidationResponse is the API response,
// containing a ownership challenge validation result.
type LogpushOwnershipChallengeValidationResponse struct {
	Response
	Result struct {
		Valid bool `json:"valid"`
	}
}

// LogpushValidateOwnershipChallengeRequest is the API request for validate ownership challenge.
type LogpushValidateOwnershipChallengeRequest struct {
	DestinationConf    string `json:"destination_conf"`
	OwnershipChallenge string `json:"ownership_challenge"`
}

// LogpushDestinationExistsResponse is the API response,
// containing a destination exists check result.
type LogpushDestinationExistsResponse struct {
	Response
	Result struct {
		Exists bool `json:"exists"`
	}
}

// LogpushDestinationExistsRequest is the API request for check destination exists.
type LogpushDestinationExistsRequest struct {
	DestinationConf string `json:"destination_conf"`
}

// Custom Marshaller for LogpushJob filter key.
func (f LogpushJob) MarshalJSON() ([]byte, error) {
	type Alias LogpushJob

	var filter string

	if f.Filter != nil {
		b, err := json.Marshal(f.Filter)

		if err != nil {
			return nil, err
		}

		filter = string(b)
	}

	return json.Marshal(&struct {
		Filter string `json:"filter,omitempty"`
		Alias
	}{
		Filter: filter,
		Alias:  (Alias)(f),
	})
}

// Custom Unmarshaller for LogpushJob filter key.
func (f *LogpushJob) UnmarshalJSON(data []byte) error {
	type Alias LogpushJob
	aux := &struct {
		Filter string `json:"filter,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(f),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux != nil && aux.Filter != "" {
		var filter LogpushJobFilters
		if err := json.Unmarshal([]byte(aux.Filter), &filter); err != nil {
			return err
		}
		if err := filter.Where.Validate(); err != nil {
			return err
		}
		f.Filter = &filter
	}
	return nil
}

func (f CreateLogpushJobParams) MarshalJSON() ([]byte, error) {
	type Alias CreateLogpushJobParams

	var filter string

	if f.Filter != nil {
		b, err := json.Marshal(f.Filter)

		if err != nil {
			return nil, err
		}

		filter = string(b)
	}

	return json.Marshal(&struct {
		Filter string `json:"filter,omitempty"`
		Alias
	}{
		Filter: filter,
		Alias:  (Alias)(f),
	})
}

// Custom Unmarshaller for CreateLogpushJobParams filter key.
func (f *CreateLogpushJobParams) UnmarshalJSON(data []byte) error {
	type Alias CreateLogpushJobParams
	aux := &struct {
		Filter string `json:"filter,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(f),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux != nil && aux.Filter != "" {
		var filter LogpushJobFilters
		if err := json.Unmarshal([]byte(aux.Filter), &filter); err != nil {
			return err
		}
		if err := filter.Where.Validate(); err != nil {
			return err
		}
		f.Filter = &filter
	}
	return nil
}

func (f UpdateLogpushJobParams) MarshalJSON() ([]byte, error) {
	type Alias UpdateLogpushJobParams

	var filter string

	if f.Filter != nil {
		b, err := json.Marshal(f.Filter)

		if err != nil {
			return nil, err
		}

		filter = string(b)
	}

	return json.Marshal(&struct {
		Filter string `json:"filter,omitempty"`
		Alias
	}{
		Filter: filter,
		Alias:  (Alias)(f),
	})
}

// Custom Unmarshaller for UpdateLogpushJobParams filter key.
func (f *UpdateLogpushJobParams) UnmarshalJSON(data []byte) error {
	type Alias UpdateLogpushJobParams
	aux := &struct {
		Filter string `json:"filter,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(f),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux != nil && aux.Filter != "" {
		var filter LogpushJobFilters
		if err := json.Unmarshal([]byte(aux.Filter), &filter); err != nil {
			return err
		}
		if err := filter.Where.Validate(); err != nil {
			return err
		}
		f.Filter = &filter
	}
	return nil
}

func (filter *LogpushJobFilter) Validate() error {
	if filter.And != nil {
		if filter.Or != nil || filter.Key != "" || filter.Operator != "" || filter.Value != nil {
			return errors.New("And can't be set with Or, Key, Operator or Value")
		}
		for i, element := range filter.And {
			err := element.Validate()
			if err != nil {
				return fmt.Errorf("element %v in And is invalid: %w", i, err)
			}
		}
		return nil
	}
	if filter.Or != nil {
		if filter.And != nil || filter.Key != "" || filter.Operator != "" || filter.Value != nil {
			return errors.New("Or can't be set with And, Key, Operator or Value")
		}
		for i, element := range filter.Or {
			err := element.Validate()
			if err != nil {
				return fmt.Errorf("element %v in Or is invalid: %w", i, err)
			}
		}
		return nil
	}
	if filter.Key == "" {
		return errors.New("Key is missing")
	}

	if filter.Operator == "" {
		return errors.New("Operator is missing")
	}

	if filter.Value == nil {
		return errors.New("Value is missing")
	}

	return nil
}

type CreateLogpushJobParams struct {
	Dataset                  string                `json:"dataset"`
	Enabled                  bool                  `json:"enabled"`
	Kind                     string                `json:"kind,omitempty"`
	Name                     string                `json:"name"`
	LogpullOptions           string                `json:"logpull_options,omitempty"`
	OutputOptions            *LogpushOutputOptions `json:"output_options,omitempty"`
	DestinationConf          string                `json:"destination_conf"`
	OwnershipChallenge       string                `json:"ownership_challenge,omitempty"`
	ErrorMessage             string                `json:"error_message,omitempty"`
	Frequency                string                `json:"frequency,omitempty"`
	Filter                   *LogpushJobFilters    `json:"filter,omitempty"`
	MaxUploadBytes           int                   `json:"max_upload_bytes,omitempty"`
	MaxUploadRecords         int                   `json:"max_upload_records,omitempty"`
	MaxUploadIntervalSeconds int                   `json:"max_upload_interval_seconds,omitempty"`
}

type ListLogpushJobsParams struct{}

type ListLogpushJobsForDatasetParams struct {
	Dataset string `json:"-"`
}

type GetLogpushFieldsParams struct {
	Dataset string `json:"-"`
}

type UpdateLogpushJobParams struct {
	ID                       int                   `json:"-"`
	Dataset                  string                `json:"dataset"`
	Enabled                  bool                  `json:"enabled"`
	Kind                     string                `json:"kind,omitempty"`
	Name                     string                `json:"name"`
	LogpullOptions           string                `json:"logpull_options,omitempty"`
	OutputOptions            *LogpushOutputOptions `json:"output_options,omitempty"`
	DestinationConf          string                `json:"destination_conf"`
	OwnershipChallenge       string                `json:"ownership_challenge,omitempty"`
	LastComplete             *time.Time            `json:"last_complete,omitempty"`
	LastError                *time.Time            `json:"last_error,omitempty"`
	ErrorMessage             string                `json:"error_message,omitempty"`
	Frequency                string                `json:"frequency,omitempty"`
	Filter                   *LogpushJobFilters    `json:"filter,omitempty"`
	MaxUploadBytes           int                   `json:"max_upload_bytes,omitempty"`
	MaxUploadRecords         int                   `json:"max_upload_records,omitempty"`
	MaxUploadIntervalSeconds int                   `json:"max_upload_interval_seconds,omitempty"`
}

type ValidateLogpushOwnershipChallengeParams struct {
	DestinationConf    string `json:"destination_conf"`
	OwnershipChallenge string `json:"ownership_challenge"`
}

type GetLogpushOwnershipChallengeParams struct {
	DestinationConf string `json:"destination_conf"`
}

// CreateLogpushJob creates a new zone-level Logpush Job.
//
// API reference: https://api.cloudflare.com/#logpush-jobs-create-logpush-job
func (api *API) CreateLogpushJob(ctx context.Context, rc *ResourceContainer, params CreateLogpushJobParams) (*LogpushJob, error) {
	uri := fmt.Sprintf("/%s/%s/logpush/jobs", rc.Level, rc.Identifier)
	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, params)
	if err != nil {
		return nil, err
	}
	var r LogpushJobDetailsResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}
	return &r.Result, nil
}

// ListAccountLogpushJobs returns all Logpush Jobs for all datasets.
//
// API reference: https://api.cloudflare.com/#logpush-jobs-list-logpush-jobs
func (api *API) ListLogpushJobs(ctx context.Context, rc *ResourceContainer, params ListLogpushJobsParams) ([]LogpushJob, error) {
	uri := fmt.Sprintf("/%s/%s/logpush/jobs", rc.Level, rc.Identifier)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []LogpushJob{}, err
	}
	var r LogpushJobsResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return []LogpushJob{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}
	return r.Result, nil
}

// LogpushJobsForDataset returns all Logpush Jobs for a dataset.
//
// API reference: https://api.cloudflare.com/#logpush-jobs-list-logpush-jobs-for-a-dataset
func (api *API) ListLogpushJobsForDataset(ctx context.Context, rc *ResourceContainer, params ListLogpushJobsForDatasetParams) ([]LogpushJob, error) {
	uri := fmt.Sprintf("/%s/%s/logpush/datasets/%s/jobs", rc.Level, rc.Identifier, params.Dataset)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []LogpushJob{}, err
	}
	var r LogpushJobsResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return []LogpushJob{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}
	return r.Result, nil
}

// LogpushFields returns fields for a given dataset.
//
// API reference: https://api.cloudflare.com/#logpush-jobs-list-logpush-jobs
func (api *API) GetLogpushFields(ctx context.Context, rc *ResourceContainer, params GetLogpushFieldsParams) (LogpushFields, error) {
	uri := fmt.Sprintf("/%s/%s/logpush/datasets/%s/fields", rc.Level, rc.Identifier, params.Dataset)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return LogpushFields{}, err
	}
	var r LogpushFieldsResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return LogpushFields{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}
	return r.Result, nil
}

// LogpushJob fetches detail about one Logpush Job for a zone.
//
// API reference: https://api.cloudflare.com/#logpush-jobs-logpush-job-details
func (api *API) GetLogpushJob(ctx context.Context, rc *ResourceContainer, jobID int) (LogpushJob, error) {
	uri := fmt.Sprintf("/%s/%s/logpush/jobs/%d", rc.Level, rc.Identifier, jobID)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return LogpushJob{}, err
	}
	var r LogpushJobDetailsResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return LogpushJob{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}
	return r.Result, nil
}

// UpdateLogpushJob lets you update a Logpush Job.
//
// API reference: https://api.cloudflare.com/#logpush-jobs-update-logpush-job
func (api *API) UpdateLogpushJob(ctx context.Context, rc *ResourceContainer, params UpdateLogpushJobParams) error {
	uri := fmt.Sprintf("/%s/%s/logpush/jobs/%d", rc.Level, rc.Identifier, params.ID)
	res, err := api.makeRequestContext(ctx, http.MethodPut, uri, params)
	if err != nil {
		return err
	}
	var r LogpushJobDetailsResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return fmt.Errorf("%s: %w", errUnmarshalError, err)
	}
	return nil
}

// DeleteLogpushJob deletes a Logpush Job for a zone.
//
// API reference: https://api.cloudflare.com/#logpush-jobs-delete-logpush-job
func (api *API) DeleteLogpushJob(ctx context.Context, rc *ResourceContainer, jobID int) error {
	uri := fmt.Sprintf("/%s/%s/logpush/jobs/%d", rc.Level, rc.Identifier, jobID)
	res, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}
	var r LogpushJobDetailsResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return fmt.Errorf("%s: %w", errUnmarshalError, err)
	}
	return nil
}

// GetLogpushOwnershipChallenge returns ownership challenge.
//
// API reference: https://api.cloudflare.com/#logpush-jobs-get-ownership-challenge
func (api *API) GetLogpushOwnershipChallenge(ctx context.Context, rc *ResourceContainer, params GetLogpushOwnershipChallengeParams) (*LogpushGetOwnershipChallenge, error) {
	uri := fmt.Sprintf("/%s/%s/logpush/ownership", rc.Level, rc.Identifier)
	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, params)
	if err != nil {
		return nil, err
	}
	var r LogpushGetOwnershipChallengeResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	if !r.Result.Valid {
		return nil, errors.New(r.Result.Message)
	}

	return &r.Result, nil
}

// ValidateLogpushOwnershipChallenge returns zone-level ownership challenge validation result.
//
// API reference: https://api.cloudflare.com/#logpush-jobs-validate-ownership-challenge
func (api *API) ValidateLogpushOwnershipChallenge(ctx context.Context, rc *ResourceContainer, params ValidateLogpushOwnershipChallengeParams) (bool, error) {
	uri := fmt.Sprintf("/%s/%s/logpush/ownership/validate", rc.Level, rc.Identifier)
	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, params)
	if err != nil {
		return false, err
	}
	var r LogpushGetOwnershipChallengeResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return false, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}
	return r.Result.Valid, nil
}

// CheckLogpushDestinationExists returns zone-level destination exists check result.
//
// API reference: https://api.cloudflare.com/#logpush-jobs-check-destination-exists
func (api *API) CheckLogpushDestinationExists(ctx context.Context, rc *ResourceContainer, destinationConf string) (bool, error) {
	uri := fmt.Sprintf("/%s/%s/logpush/validate/destination/exists", rc.Level, rc.Identifier)
	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, LogpushDestinationExistsRequest{
		DestinationConf: destinationConf,
	})
	if err != nil {
		return false, err
	}
	var r LogpushDestinationExistsResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return false, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}
	return r.Result.Exists, nil
}
