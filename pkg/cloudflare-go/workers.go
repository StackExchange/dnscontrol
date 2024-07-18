package cloudflare

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"github.com/goccy/go-json"
)

// WorkerRequestParams provides parameters for worker requests for both enterprise and standard requests.
type WorkerRequestParams struct {
	ZoneID     string
	ScriptName string
}

type CreateWorkerParams struct {
	ScriptName string
	Script     string

	// DispatchNamespaceName uploads the worker to a WFP dispatch namespace if provided
	DispatchNamespaceName *string

	// Module changes the Content-Type header to specify the script is an
	// ES Module syntax script.
	Module bool

	// Logpush opts the worker into Workers Logpush logging. A nil value leaves
	// the current setting unchanged.
	//
	// Documentation: https://developers.cloudflare.com/workers/platform/logpush/
	Logpush *bool

	// TailConsumers specifies a list of Workers that will consume the logs of
	// the attached Worker.
	// Documentation: https://developers.cloudflare.com/workers/platform/tail-workers/
	TailConsumers *[]WorkersTailConsumer

	// Bindings should be a map where the keys are the binding name, and the
	// values are the binding content
	Bindings map[string]WorkerBinding

	// CompatibilityDate is a date in the form yyyy-mm-dd,
	// which will be used to determine which version of the Workers runtime is used.
	//  https://developers.cloudflare.com/workers/platform/compatibility-dates/
	CompatibilityDate string

	// CompatibilityFlags are the names of features of the Workers runtime to be enabled or disabled,
	// usually used together with CompatibilityDate.
	//  https://developers.cloudflare.com/workers/platform/compatibility-dates/#compatibility-flags
	CompatibilityFlags []string

	Placement *Placement

	// Tags are used to better manage CRUD operations at scale.
	//  https://developers.cloudflare.com/cloudflare-for-platforms/workers-for-platforms/platform/tags/
	Tags []string
}

func (p CreateWorkerParams) RequiresMultipart() bool {
	switch {
	case p.Module:
		return true
	case p.Logpush != nil:
		return true
	case p.Placement != nil:
		return true
	case len(p.Bindings) > 0:
		return true
	case p.CompatibilityDate != "":
		return true
	case len(p.CompatibilityFlags) > 0:
		return true
	case p.TailConsumers != nil:
		return true
	case len(p.Tags) > 0:
		return true
	}

	return false
}

type UpdateWorkersScriptContentParams struct {
	ScriptName string
	Script     string

	// DispatchNamespaceName uploads the worker to a WFP dispatch namespace if provided
	DispatchNamespaceName *string

	// Module changes the Content-Type header to specify the script is an
	// ES Module syntax script.
	Module bool
}

type UpdateWorkersScriptSettingsParams struct {
	ScriptName string

	// Logpush opts the worker into Workers Logpush logging. A nil value leaves
	// the current setting unchanged.
	//
	// Documentation: https://developers.cloudflare.com/workers/platform/logpush/
	Logpush *bool

	// TailConsumers specifies a list of Workers that will consume the logs of
	// the attached Worker.
	// Documentation: https://developers.cloudflare.com/workers/platform/tail-workers/
	TailConsumers *[]WorkersTailConsumer

	// Bindings should be a map where the keys are the binding name, and the
	// values are the binding content
	Bindings map[string]WorkerBinding

	// CompatibilityDate is a date in the form yyyy-mm-dd,
	// which will be used to determine which version of the Workers runtime is used.
	//  https://developers.cloudflare.com/workers/platform/compatibility-dates/
	CompatibilityDate string

	// CompatibilityFlags are the names of features of the Workers runtime to be enabled or disabled,
	// usually used together with CompatibilityDate.
	//  https://developers.cloudflare.com/workers/platform/compatibility-dates/#compatibility-flags
	CompatibilityFlags []string

	Placement *Placement
}

// WorkerScriptParams provides a worker script and the associated bindings.
type WorkerScriptParams struct {
	ScriptName string

	// Module changes the Content-Type header to specify the script is an
	// ES Module syntax script.
	Module bool

	// Bindings should be a map where the keys are the binding name, and the
	// values are the binding content
	Bindings map[string]WorkerBinding
}

// WorkerRoute is used to map traffic matching a URL pattern to a workers
//
// API reference: https://api.cloudflare.com/#worker-routes-properties
type WorkerRoute struct {
	ID         string `json:"id,omitempty"`
	Pattern    string `json:"pattern"`
	ScriptName string `json:"script,omitempty"`
}

// WorkerRoutesResponse embeds Response struct and slice of WorkerRoutes.
type WorkerRoutesResponse struct {
	Response
	Routes []WorkerRoute `json:"result"`
}

// WorkerRouteResponse embeds Response struct and a single WorkerRoute.
type WorkerRouteResponse struct {
	Response
	WorkerRoute `json:"result"`
}

// WorkerScript Cloudflare Worker struct with metadata.
type WorkerScript struct {
	WorkerMetaData
	Script     string `json:"script"`
	UsageModel string `json:"usage_model,omitempty"`
}

type WorkersTailConsumer struct {
	Service     string  `json:"service"`
	Environment *string `json:"environment,omitempty"`
	Namespace   *string `json:"namespace,omitempty"`
}

// WorkerMetaData contains worker script information such as size, creation & modification dates.
type WorkerMetaData struct {
	ID               string                 `json:"id,omitempty"`
	ETAG             string                 `json:"etag,omitempty"`
	Size             int                    `json:"size,omitempty"`
	CreatedOn        time.Time              `json:"created_on,omitempty"`
	ModifiedOn       time.Time              `json:"modified_on,omitempty"`
	Logpush          *bool                  `json:"logpush,omitempty"`
	TailConsumers    *[]WorkersTailConsumer `json:"tail_consumers,omitempty"`
	LastDeployedFrom *string                `json:"last_deployed_from,omitempty"`
	DeploymentId     *string                `json:"deployment_id,omitempty"`
	PlacementMode    *PlacementMode         `json:"placement_mode,omitempty"`
	PipelineHash     *string                `json:"pipeline_hash,omitempty"`
}

// WorkerListResponse wrapper struct for API response to worker script list API call.
type WorkerListResponse struct {
	Response
	ResultInfo
	WorkerList []WorkerMetaData `json:"result"`
}

// WorkerScriptResponse wrapper struct for API response to worker script calls.
type WorkerScriptResponse struct {
	Response
	Module       bool
	WorkerScript `json:"result"`
}

// WorkerScriptSettingsResponse wrapper struct for API response to worker script settings calls.
type WorkerScriptSettingsResponse struct {
	Response
	WorkerMetaData
}

type ListWorkersParams struct{}

type DeleteWorkerParams struct {
	ScriptName string

	// DispatchNamespaceName is the dispatch namespace the Worker is uploaded to.
	DispatchNamespace *string
}

type PlacementMode string

const (
	PlacementModeOff   PlacementMode = ""
	PlacementModeSmart PlacementMode = "smart"
)

type Placement struct {
	Mode PlacementMode `json:"mode"`
}

// DeleteWorker deletes a single Worker.
//
// API reference: https://developers.cloudflare.com/api/operations/worker-script-delete-worker
func (api *API) DeleteWorker(ctx context.Context, rc *ResourceContainer, params DeleteWorkerParams) error {
	if rc.Level != AccountRouteLevel {
		return ErrRequiredAccountLevelResourceContainer
	}

	if rc.Identifier == "" {
		return ErrMissingAccountID
	}

	uri := fmt.Sprintf("/accounts/%s/workers/scripts/%s", rc.Identifier, params.ScriptName)
	if params.DispatchNamespace != nil && *params.DispatchNamespace != "" {
		uri = fmt.Sprintf("/accounts/%s/workers/dispatch/namespaces/%s/scripts/%s", rc.Identifier, *params.DispatchNamespace, params.ScriptName)
	}

	res, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)

	var r WorkerScriptResponse
	if err != nil {
		return err
	}

	err = json.Unmarshal(res, &r)
	if err != nil {
		return fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return nil
}

// GetWorker fetch raw script content for your worker returns string containing
// worker code js.
//
// API reference: https://developers.cloudflare.com/api/operations/worker-script-download-worker
func (api *API) GetWorker(ctx context.Context, rc *ResourceContainer, scriptName string) (WorkerScriptResponse, error) {
	return api.GetWorkerWithDispatchNamespace(ctx, rc, scriptName, "")
}

// GetWorker fetch raw script content for your worker returns string containing
// worker code js.
//
// API reference: https://developers.cloudflare.com/api/operations/worker-script-download-worker
func (api *API) GetWorkerWithDispatchNamespace(ctx context.Context, rc *ResourceContainer, scriptName string, dispatchNamespace string) (WorkerScriptResponse, error) {
	if rc.Level != AccountRouteLevel {
		return WorkerScriptResponse{}, ErrRequiredAccountLevelResourceContainer
	}

	if rc.Identifier == "" {
		return WorkerScriptResponse{}, ErrMissingAccountID
	}

	uri := fmt.Sprintf("/accounts/%s/workers/scripts/%s", rc.Identifier, scriptName)
	if dispatchNamespace != "" {
		uri = fmt.Sprintf("/accounts/%s/workers/dispatch/namespaces/%s/scripts/%s/content", rc.Identifier, dispatchNamespace, scriptName)
	}
	res, err := api.makeRequestContextWithHeadersComplete(ctx, http.MethodGet, uri, nil, nil)
	var r WorkerScriptResponse
	if err != nil {
		return r, err
	}

	// Check if the response type is multipart, in which case this was a module worker
	mediaType, mediaParams, _ := mime.ParseMediaType(res.Headers.Get("content-type"))
	if strings.HasPrefix(mediaType, "multipart/") {
		bytesReader := bytes.NewReader(res.Body)
		mimeReader := multipart.NewReader(bytesReader, mediaParams["boundary"])
		mimePart, err := mimeReader.NextPart()
		if err != nil {
			return r, fmt.Errorf("could not get multipart response body: %w", err)
		}
		mimePartBody, err := io.ReadAll(mimePart)
		if err != nil {
			return r, fmt.Errorf("could not read multipart response body: %w", err)
		}
		r.Script = string(mimePartBody)
		r.Module = true
	} else {
		r.Script = string(res.Body)
		r.Module = false
	}

	r.Success = true
	return r, nil
}

// ListWorkers returns list of Workers for given account.
//
// API reference: https://developers.cloudflare.com/api/operations/worker-script-list-workers
func (api *API) ListWorkers(ctx context.Context, rc *ResourceContainer, params ListWorkersParams) (WorkerListResponse, *ResultInfo, error) {
	if rc.Level != AccountRouteLevel {
		return WorkerListResponse{}, &ResultInfo{}, ErrRequiredAccountLevelResourceContainer
	}

	if rc.Identifier == "" {
		return WorkerListResponse{}, &ResultInfo{}, ErrMissingAccountID
	}

	uri := fmt.Sprintf("/accounts/%s/workers/scripts", rc.Identifier)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return WorkerListResponse{}, &ResultInfo{}, err
	}

	var r WorkerListResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return WorkerListResponse{}, &ResultInfo{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return r, &r.ResultInfo, nil
}

// UploadWorker pushes raw script content for your Worker.
//
// API reference: https://developers.cloudflare.com/api/operations/worker-script-upload-worker-module
func (api *API) UploadWorker(ctx context.Context, rc *ResourceContainer, params CreateWorkerParams) (WorkerScriptResponse, error) {
	if rc.Level != AccountRouteLevel {
		return WorkerScriptResponse{}, ErrRequiredAccountLevelResourceContainer
	}

	if rc.Identifier == "" {
		return WorkerScriptResponse{}, ErrMissingAccountID
	}

	body := []byte(params.Script)
	var (
		contentType = "application/javascript"
		err         error
	)

	if params.RequiresMultipart() {
		contentType, body, err = formatMultipartBody(params)
		if err != nil {
			return WorkerScriptResponse{}, err
		}
	}

	uri := fmt.Sprintf("/accounts/%s/workers/scripts/%s", rc.Identifier, params.ScriptName)
	if params.DispatchNamespaceName != nil && *params.DispatchNamespaceName != "" {
		uri = fmt.Sprintf("/accounts/%s/workers/dispatch/namespaces/%s/scripts/%s", rc.Identifier, *params.DispatchNamespaceName, params.ScriptName)
	}

	headers := make(http.Header)
	headers.Set("Content-Type", contentType)
	res, err := api.makeRequestContextWithHeaders(ctx, http.MethodPut, uri, body, headers)

	var r WorkerScriptResponse
	if err != nil {
		return r, err
	}

	err = json.Unmarshal(res, &r)
	if err != nil {
		return r, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return r, nil
}

// GetWorkersScriptContent returns the pure script content of a worker.
//
// API reference: https://developers.cloudflare.com/api/operations/worker-script-get-content
func (api *API) GetWorkersScriptContent(ctx context.Context, rc *ResourceContainer, scriptName string) (string, error) {
	if rc.Level != AccountRouteLevel {
		return "", ErrRequiredAccountLevelResourceContainer
	}

	if rc.Identifier == "" {
		return "", ErrMissingAccountID
	}

	uri := fmt.Sprintf("/accounts/%s/workers/scripts/%s/content/v2", rc.Identifier, scriptName)
	res, err := api.makeRequestContextWithHeadersComplete(ctx, http.MethodGet, uri, nil, nil)
	if err != nil {
		return "", err
	}

	return string(res.Body), nil
}

// UpdateWorkersScriptContent pushes only script content, no metadata.
//
// API reference: https://developers.cloudflare.com/api/operations/worker-script-put-content
func (api *API) UpdateWorkersScriptContent(ctx context.Context, rc *ResourceContainer, params UpdateWorkersScriptContentParams) (WorkerScriptResponse, error) {
	if rc.Level != AccountRouteLevel {
		return WorkerScriptResponse{}, ErrRequiredAccountLevelResourceContainer
	}

	if rc.Identifier == "" {
		return WorkerScriptResponse{}, ErrMissingAccountID
	}

	body := []byte(params.Script)
	var (
		contentType = "application/javascript"
		err         error
	)

	if params.Module {
		var formattedParams CreateWorkerParams
		formattedParams.Script = params.Script
		formattedParams.ScriptName = params.ScriptName
		formattedParams.Module = params.Module
		formattedParams.DispatchNamespaceName = params.DispatchNamespaceName
		contentType, body, err = formatMultipartBody(formattedParams)
		if err != nil {
			return WorkerScriptResponse{}, err
		}
	}

	uri := fmt.Sprintf("/accounts/%s/workers/scripts/%s/content", rc.Identifier, params.ScriptName)
	if params.DispatchNamespaceName != nil {
		uri = fmt.Sprintf("/accounts/%s/workers/dispatch_namespaces/%s/scripts/%s/content", rc.Identifier, *params.DispatchNamespaceName, params.ScriptName)
	}

	headers := make(http.Header)
	headers.Set("Content-Type", contentType)
	res, err := api.makeRequestContextWithHeaders(ctx, http.MethodPut, uri, body, headers)

	var r WorkerScriptResponse
	if err != nil {
		return r, err
	}

	err = json.Unmarshal(res, &r)
	if err != nil {
		return r, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return r, nil
}

// GetWorkersScriptSettings returns the metadata of a worker.
//
// API reference: https://developers.cloudflare.com/api/operations/worker-script-get-settings
func (api *API) GetWorkersScriptSettings(ctx context.Context, rc *ResourceContainer, scriptName string) (WorkerScriptSettingsResponse, error) {
	if rc.Level != AccountRouteLevel {
		return WorkerScriptSettingsResponse{}, ErrRequiredAccountLevelResourceContainer
	}

	if rc.Identifier == "" {
		return WorkerScriptSettingsResponse{}, ErrMissingAccountID
	}

	uri := fmt.Sprintf("/accounts/%s/workers/scripts/%s/settings", rc.Identifier, scriptName)
	res, err := api.makeRequestContextWithHeaders(ctx, http.MethodGet, uri, nil, nil)
	var r WorkerScriptSettingsResponse
	if err != nil {
		return r, err
	}

	err = json.Unmarshal(res, &r)
	if err != nil {
		return r, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	r.Success = true

	return r, nil
}

// UpdateWorkersScriptSettings pushes only script metadata.
//
// API reference: https://developers.cloudflare.com/api/operations/worker-script-patch-settings
func (api *API) UpdateWorkersScriptSettings(ctx context.Context, rc *ResourceContainer, params UpdateWorkersScriptSettingsParams) (WorkerScriptSettingsResponse, error) {
	if rc.Level != AccountRouteLevel {
		return WorkerScriptSettingsResponse{}, ErrRequiredAccountLevelResourceContainer
	}

	if rc.Identifier == "" {
		return WorkerScriptSettingsResponse{}, ErrMissingAccountID
	}

	body, err := json.Marshal(params)
	if err != nil {
		return WorkerScriptSettingsResponse{}, err
	}
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")

	uri := fmt.Sprintf("/accounts/%s/workers/scripts/%s/settings", rc.Identifier, params.ScriptName)
	res, err := api.makeRequestContextWithHeaders(ctx, http.MethodPatch, uri, body, headers)
	var r WorkerScriptSettingsResponse
	if err != nil {
		return r, err
	}

	err = json.Unmarshal(res, &r)
	if err != nil {
		return r, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	r.Success = true

	return r, nil
}

// Returns content-type, body, error.
func formatMultipartBody(params CreateWorkerParams) (string, []byte, error) {
	var buf = &bytes.Buffer{}
	var mpw = multipart.NewWriter(buf)
	defer mpw.Close()

	// Write metadata part
	var scriptPartName string
	meta := struct {
		BodyPart           string                 `json:"body_part,omitempty"`
		MainModule         string                 `json:"main_module,omitempty"`
		Bindings           []workerBindingMeta    `json:"bindings"`
		Logpush            *bool                  `json:"logpush,omitempty"`
		TailConsumers      *[]WorkersTailConsumer `json:"tail_consumers,omitempty"`
		CompatibilityDate  string                 `json:"compatibility_date,omitempty"`
		CompatibilityFlags []string               `json:"compatibility_flags,omitempty"`
		Placement          *Placement             `json:"placement,omitempty"`
		Tags               []string               `json:"tags"`
	}{
		Bindings:           make([]workerBindingMeta, 0, len(params.Bindings)),
		Logpush:            params.Logpush,
		TailConsumers:      params.TailConsumers,
		CompatibilityDate:  params.CompatibilityDate,
		CompatibilityFlags: params.CompatibilityFlags,
		Placement:          params.Placement,
		Tags:               params.Tags,
	}

	if params.Module {
		scriptPartName = "worker.mjs"
		meta.MainModule = scriptPartName
	} else {
		scriptPartName = "script"
		meta.BodyPart = scriptPartName
	}

	bodyWriters := make([]workerBindingBodyWriter, 0, len(params.Bindings))
	for name, b := range params.Bindings {
		bindingMeta, bodyWriter, err := b.serialize(name)
		if err != nil {
			return "", nil, err
		}

		meta.Bindings = append(meta.Bindings, bindingMeta)
		bodyWriters = append(bodyWriters, bodyWriter)
	}

	var hdr = textproto.MIMEHeader{}
	hdr.Set("content-disposition", fmt.Sprintf(`form-data; name="%s"`, "metadata"))
	hdr.Set("content-type", "application/json")
	pw, err := mpw.CreatePart(hdr)
	if err != nil {
		return "", nil, err
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return "", nil, err
	}
	_, err = pw.Write(metaJSON)
	if err != nil {
		return "", nil, err
	}

	// Write script part
	hdr = textproto.MIMEHeader{}

	contentType := "application/javascript"
	if params.Module {
		contentType = "application/javascript+module"
		hdr.Set("content-disposition", fmt.Sprintf(`form-data; name="%s"; filename="%[1]s"`, scriptPartName))
	} else {
		hdr.Set("content-disposition", fmt.Sprintf(`form-data; name="%s"`, scriptPartName))
	}
	hdr.Set("content-type", contentType)

	pw, err = mpw.CreatePart(hdr)
	if err != nil {
		return "", nil, err
	}
	_, err = pw.Write([]byte(params.Script))
	if err != nil {
		return "", nil, err
	}

	// Write other bindings with parts
	for _, w := range bodyWriters {
		if w != nil {
			err = w(mpw)
			if err != nil {
				return "", nil, err
			}
		}
	}

	mpw.Close()

	return mpw.FormDataContentType(), buf.Bytes(), nil
}
