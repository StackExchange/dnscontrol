package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/goccy/go-json"
)

var (
	ErrMissingRulesetPhase = errors.New("missing required phase")
)

const (
	RulesetKindCustom  RulesetKind = "custom"
	RulesetKindManaged RulesetKind = "managed"
	RulesetKindRoot    RulesetKind = "root"
	RulesetKindZone    RulesetKind = "zone"

	RulesetPhaseDDoSL4                       RulesetPhase = "ddos_l4"
	RulesetPhaseDDoSL7                       RulesetPhase = "ddos_l7"
	RulesetPhaseHTTPConfigSettings           RulesetPhase = "http_config_settings"
	RulesetPhaseHTTPCustomErrors             RulesetPhase = "http_custom_errors"
	RulesetPhaseHTTPLogCustomFields          RulesetPhase = "http_log_custom_fields"
	RulesetPhaseHTTPRatelimit                RulesetPhase = "http_ratelimit"
	RulesetPhaseHTTPRequestCacheSettings     RulesetPhase = "http_request_cache_settings"
	RulesetPhaseHTTPRequestDynamicRedirect   RulesetPhase = "http_request_dynamic_redirect" //nolint:gosec
	RulesetPhaseHTTPRequestFirewallCustom    RulesetPhase = "http_request_firewall_custom"
	RulesetPhaseHTTPRequestFirewallManaged   RulesetPhase = "http_request_firewall_managed"
	RulesetPhaseHTTPRequestLateTransform     RulesetPhase = "http_request_late_transform"
	RulesetPhaseHTTPRequestOrigin            RulesetPhase = "http_request_origin"
	RulesetPhaseHTTPRequestRedirect          RulesetPhase = "http_request_redirect"
	RulesetPhaseHTTPRequestSanitize          RulesetPhase = "http_request_sanitize"
	RulesetPhaseHTTPRequestSBFM              RulesetPhase = "http_request_sbfm"
	RulesetPhaseHTTPRequestTransform         RulesetPhase = "http_request_transform"
	RulesetPhaseHTTPResponseCompression      RulesetPhase = "http_response_compression"
	RulesetPhaseHTTPResponseFirewallManaged  RulesetPhase = "http_response_firewall_managed"
	RulesetPhaseHTTPResponseHeadersTransform RulesetPhase = "http_response_headers_transform"
	RulesetPhaseMagicTransit                 RulesetPhase = "magic_transit"

	RulesetRuleActionBlock                RulesetRuleAction = "block"
	RulesetRuleActionChallenge            RulesetRuleAction = "challenge"
	RulesetRuleActionCompressResponse     RulesetRuleAction = "compress_response"
	RulesetRuleActionDDoSDynamic          RulesetRuleAction = "ddos_dynamic"
	RulesetRuleActionDDoSMitigation       RulesetRuleAction = "ddos_mitigation"
	RulesetRuleActionExecute              RulesetRuleAction = "execute"
	RulesetRuleActionForceConnectionClose RulesetRuleAction = "force_connection_close"
	RulesetRuleActionJSChallenge          RulesetRuleAction = "js_challenge"
	RulesetRuleActionLog                  RulesetRuleAction = "log"
	RulesetRuleActionLogCustomField       RulesetRuleAction = "log_custom_field"
	RulesetRuleActionManagedChallenge     RulesetRuleAction = "managed_challenge"
	RulesetRuleActionRedirect             RulesetRuleAction = "redirect"
	RulesetRuleActionRewrite              RulesetRuleAction = "rewrite"
	RulesetRuleActionRoute                RulesetRuleAction = "route"
	RulesetRuleActionScore                RulesetRuleAction = "score"
	RulesetRuleActionServeError           RulesetRuleAction = "serve_error"
	RulesetRuleActionSetCacheSettings     RulesetRuleAction = "set_cache_settings"
	RulesetRuleActionSetConfig            RulesetRuleAction = "set_config"
	RulesetRuleActionSkip                 RulesetRuleAction = "skip"

	RulesetActionParameterProductBIC           RulesetActionParameterProduct = "bic"
	RulesetActionParameterProductHOT           RulesetActionParameterProduct = "hot"
	RulesetActionParameterProductRateLimit     RulesetActionParameterProduct = "ratelimit"
	RulesetActionParameterProductSecurityLevel RulesetActionParameterProduct = "securityLevel"
	RulesetActionParameterProductUABlock       RulesetActionParameterProduct = "uablock"
	RulesetActionParameterProductWAF           RulesetActionParameterProduct = "waf"
	RulesetActionParameterProductZoneLockdown  RulesetActionParameterProduct = "zonelockdown"

	RulesetRuleActionParametersHTTPHeaderOperationRemove RulesetRuleActionParametersHTTPHeaderOperation = "remove"
	RulesetRuleActionParametersHTTPHeaderOperationSet    RulesetRuleActionParametersHTTPHeaderOperation = "set"
	RulesetRuleActionParametersHTTPHeaderOperationAdd    RulesetRuleActionParametersHTTPHeaderOperation = "add"
)

// RulesetKindValues exposes all the available `RulesetKind` values as a slice
// of strings.
func RulesetKindValues() []string {
	return []string{
		string(RulesetKindCustom),
		string(RulesetKindManaged),
		string(RulesetKindRoot),
		string(RulesetKindZone),
	}
}

// RulesetPhaseValues exposes all the available `RulesetPhase` values as a slice
// of strings.
func RulesetPhaseValues() []string {
	return []string{
		string(RulesetPhaseDDoSL4),
		string(RulesetPhaseDDoSL7),
		string(RulesetPhaseHTTPConfigSettings),
		string(RulesetPhaseHTTPCustomErrors),
		string(RulesetPhaseHTTPLogCustomFields),
		string(RulesetPhaseHTTPRatelimit),
		string(RulesetPhaseHTTPRequestCacheSettings),
		string(RulesetPhaseHTTPRequestDynamicRedirect),
		string(RulesetPhaseHTTPRequestFirewallCustom),
		string(RulesetPhaseHTTPRequestFirewallManaged),
		string(RulesetPhaseHTTPRequestLateTransform),
		string(RulesetPhaseHTTPRequestOrigin),
		string(RulesetPhaseHTTPRequestRedirect),
		string(RulesetPhaseHTTPRequestSanitize),
		string(RulesetPhaseHTTPRequestSBFM),
		string(RulesetPhaseHTTPRequestTransform),
		string(RulesetPhaseHTTPResponseCompression),
		string(RulesetPhaseHTTPResponseFirewallManaged),
		string(RulesetPhaseHTTPResponseHeadersTransform),
		string(RulesetPhaseMagicTransit),
	}
}

// RulesetRuleActionValues exposes all the available `RulesetRuleAction` values
// as a slice of strings.
func RulesetRuleActionValues() []string {
	return []string{
		string(RulesetRuleActionBlock),
		string(RulesetRuleActionChallenge),
		string(RulesetRuleActionCompressResponse),
		string(RulesetRuleActionDDoSDynamic),
		string(RulesetRuleActionDDoSMitigation),
		string(RulesetRuleActionExecute),
		string(RulesetRuleActionForceConnectionClose),
		string(RulesetRuleActionJSChallenge),
		string(RulesetRuleActionLog),
		string(RulesetRuleActionLogCustomField),
		string(RulesetRuleActionManagedChallenge),
		string(RulesetRuleActionRedirect),
		string(RulesetRuleActionRewrite),
		string(RulesetRuleActionRoute),
		string(RulesetRuleActionScore),
		string(RulesetRuleActionServeError),
		string(RulesetRuleActionSetCacheSettings),
		string(RulesetRuleActionSetConfig),
		string(RulesetRuleActionSkip),
	}
}

// RulesetActionParameterProductValues exposes all the available
// `RulesetActionParameterProduct` values as a slice of strings.
func RulesetActionParameterProductValues() []string {
	return []string{
		string(RulesetActionParameterProductBIC),
		string(RulesetActionParameterProductHOT),
		string(RulesetActionParameterProductRateLimit),
		string(RulesetActionParameterProductSecurityLevel),
		string(RulesetActionParameterProductUABlock),
		string(RulesetActionParameterProductWAF),
		string(RulesetActionParameterProductZoneLockdown),
	}
}

func RulesetRuleActionParametersHTTPHeaderOperationValues() []string {
	return []string{
		string(RulesetRuleActionParametersHTTPHeaderOperationRemove),
		string(RulesetRuleActionParametersHTTPHeaderOperationSet),
		string(RulesetRuleActionParametersHTTPHeaderOperationAdd),
	}
}

// RulesetRuleAction defines a custom type that is used to express allowed
// values for the rule action.
type RulesetRuleAction string

// RulesetKind is the custom type for allowed variances of rulesets.
type RulesetKind string

// RulesetPhase is the custom type for defining at what point the ruleset will
// be applied in the request pipeline.
type RulesetPhase string

// RulesetActionParameterProduct is the custom type for defining what products
// can be used within the action parameters of a ruleset.
type RulesetActionParameterProduct string

// RulesetRuleActionParametersHTTPHeaderOperation defines available options for
// HTTP header operations in actions.
type RulesetRuleActionParametersHTTPHeaderOperation string

// Ruleset contains the structure of a Ruleset. Using `string` for Kind and
// Phase is a developer nicety to support downstream clients like Terraform who
// don't really have a strong and expansive type system. As always, the
// recommendation is to use the types provided where possible to avoid
// surprises.
type Ruleset struct {
	ID                       string        `json:"id,omitempty"`
	Name                     string        `json:"name,omitempty"`
	Description              string        `json:"description,omitempty"`
	Kind                     string        `json:"kind,omitempty"`
	Version                  *string       `json:"version,omitempty"`
	LastUpdated              *time.Time    `json:"last_updated,omitempty"`
	Phase                    string        `json:"phase,omitempty"`
	Rules                    []RulesetRule `json:"rules"`
	ShareableEntitlementName string        `json:"shareable_entitlement_name,omitempty"`
}

// RulesetActionParametersLogCustomField wraps an object that is part of
// request_fields, response_fields or cookie_fields.
type RulesetActionParametersLogCustomField struct {
	Name string `json:"name,omitempty"`
}

// RulesetRuleActionParameters specifies the action parameters for a Ruleset
// rule.
type RulesetRuleActionParameters struct {
	ID                       string                                            `json:"id,omitempty"`
	Ruleset                  string                                            `json:"ruleset,omitempty"`
	Rulesets                 []string                                          `json:"rulesets,omitempty"`
	Rules                    map[string][]string                               `json:"rules,omitempty"`
	Increment                int                                               `json:"increment,omitempty"`
	URI                      *RulesetRuleActionParametersURI                   `json:"uri,omitempty"`
	Headers                  map[string]RulesetRuleActionParametersHTTPHeader  `json:"headers,omitempty"`
	Products                 []string                                          `json:"products,omitempty"`
	Phases                   []string                                          `json:"phases,omitempty"`
	Overrides                *RulesetRuleActionParametersOverrides             `json:"overrides,omitempty"`
	MatchedData              *RulesetRuleActionParametersMatchedData           `json:"matched_data,omitempty"`
	Version                  *string                                           `json:"version,omitempty"`
	Response                 *RulesetRuleActionParametersBlockResponse         `json:"response,omitempty"`
	HostHeader               string                                            `json:"host_header,omitempty"`
	Origin                   *RulesetRuleActionParametersOrigin                `json:"origin,omitempty"`
	SNI                      *RulesetRuleActionParametersSni                   `json:"sni,omitempty"`
	RequestFields            []RulesetActionParametersLogCustomField           `json:"request_fields,omitempty"`
	ResponseFields           []RulesetActionParametersLogCustomField           `json:"response_fields,omitempty"`
	CookieFields             []RulesetActionParametersLogCustomField           `json:"cookie_fields,omitempty"`
	Cache                    *bool                                             `json:"cache,omitempty"`
	AdditionalCacheablePorts []int                                             `json:"additional_cacheable_ports,omitempty"`
	EdgeTTL                  *RulesetRuleActionParametersEdgeTTL               `json:"edge_ttl,omitempty"`
	BrowserTTL               *RulesetRuleActionParametersBrowserTTL            `json:"browser_ttl,omitempty"`
	ServeStale               *RulesetRuleActionParametersServeStale            `json:"serve_stale,omitempty"`
	Content                  string                                            `json:"content,omitempty"`
	ContentType              string                                            `json:"content_type,omitempty"`
	StatusCode               uint16                                            `json:"status_code,omitempty"`
	RespectStrongETags       *bool                                             `json:"respect_strong_etags,omitempty"`
	CacheKey                 *RulesetRuleActionParametersCacheKey              `json:"cache_key,omitempty"`
	OriginCacheControl       *bool                                             `json:"origin_cache_control,omitempty"`
	OriginErrorPagePassthru  *bool                                             `json:"origin_error_page_passthru,omitempty"`
	CacheReserve             *RulesetRuleActionParametersCacheReserve          `json:"cache_reserve,omitempty"`
	FromList                 *RulesetRuleActionParametersFromList              `json:"from_list,omitempty"`
	FromValue                *RulesetRuleActionParametersFromValue             `json:"from_value,omitempty"`
	AutomaticHTTPSRewrites   *bool                                             `json:"automatic_https_rewrites,omitempty"`
	AutoMinify               *RulesetRuleActionParametersAutoMinify            `json:"autominify,omitempty"`
	BrowserIntegrityCheck    *bool                                             `json:"bic,omitempty"`
	DisableApps              *bool                                             `json:"disable_apps,omitempty"`
	DisableZaraz             *bool                                             `json:"disable_zaraz,omitempty"`
	DisableRailgun           *bool                                             `json:"disable_railgun,omitempty"`
	DisableRUM               *bool                                             `json:"disable_rum,omitempty"`
	EmailObfuscation         *bool                                             `json:"email_obfuscation,omitempty"`
	Fonts                    *bool                                             `json:"fonts,omitempty"`
	Mirage                   *bool                                             `json:"mirage,omitempty"`
	OpportunisticEncryption  *bool                                             `json:"opportunistic_encryption,omitempty"`
	Polish                   *Polish                                           `json:"polish,omitempty"`
	ReadTimeout              *uint                                             `json:"read_timeout,omitempty"`
	RocketLoader             *bool                                             `json:"rocket_loader,omitempty"`
	SecurityLevel            *SecurityLevel                                    `json:"security_level,omitempty"`
	ServerSideExcludes       *bool                                             `json:"server_side_excludes,omitempty"`
	SSL                      *SSL                                              `json:"ssl,omitempty"`
	SXG                      *bool                                             `json:"sxg,omitempty"`
	HotLinkProtection        *bool                                             `json:"hotlink_protection,omitempty"`
	Algorithms               []RulesetRuleActionParametersCompressionAlgorithm `json:"algorithms,omitempty"`
}

// RulesetRuleActionParametersFromList holds the FromList struct for
// actions which pull data from a list.
type RulesetRuleActionParametersFromList struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type RulesetRuleActionParametersAutoMinify struct {
	HTML bool `json:"html"`
	CSS  bool `json:"css"`
	JS   bool `json:"js"`
}

type RulesetRuleActionParametersFromValue struct {
	StatusCode          uint16                               `json:"status_code,omitempty"`
	TargetURL           RulesetRuleActionParametersTargetURL `json:"target_url"`
	PreserveQueryString *bool                                `json:"preserve_query_string,omitempty"`
}

type RulesetRuleActionParametersTargetURL struct {
	Value      string `json:"value,omitempty"`
	Expression string `json:"expression,omitempty"`
}

type RulesetRuleActionParametersEdgeTTL struct {
	Mode          string                                     `json:"mode,omitempty"`
	Default       *uint                                      `json:"default,omitempty"`
	StatusCodeTTL []RulesetRuleActionParametersStatusCodeTTL `json:"status_code_ttl,omitempty"`
}

type RulesetRuleActionParametersStatusCodeTTL struct {
	StatusCodeRange *RulesetRuleActionParametersStatusCodeRange `json:"status_code_range,omitempty"`
	StatusCodeValue *uint                                       `json:"status_code,omitempty"`
	Value           *int                                        `json:"value,omitempty"`
}

type RulesetRuleActionParametersStatusCodeRange struct {
	From *uint `json:"from,omitempty"`
	To   *uint `json:"to,omitempty"`
}

type RulesetRuleActionParametersBrowserTTL struct {
	Mode    string `json:"mode"`
	Default *uint  `json:"default,omitempty"`
}

type RulesetRuleActionParametersServeStale struct {
	DisableStaleWhileUpdating *bool `json:"disable_stale_while_updating,omitempty"`
}

type RulesetRuleActionParametersCacheKey struct {
	CacheByDeviceType       *bool                                 `json:"cache_by_device_type,omitempty"`
	IgnoreQueryStringsOrder *bool                                 `json:"ignore_query_strings_order,omitempty"`
	CacheDeceptionArmor     *bool                                 `json:"cache_deception_armor,omitempty"`
	CustomKey               *RulesetRuleActionParametersCustomKey `json:"custom_key,omitempty"`
}

type RulesetRuleActionParametersCacheReserve struct {
	Eligible        *bool `json:"eligible,omitempty"`
	MinimumFileSize *uint `json:"minimum_file_size,omitempty"`
}

type RulesetRuleActionParametersCustomKey struct {
	Query  *RulesetRuleActionParametersCustomKeyQuery  `json:"query_string,omitempty"`
	Header *RulesetRuleActionParametersCustomKeyHeader `json:"header,omitempty"`
	Cookie *RulesetRuleActionParametersCustomKeyCookie `json:"cookie,omitempty"`
	User   *RulesetRuleActionParametersCustomKeyUser   `json:"user,omitempty"`
	Host   *RulesetRuleActionParametersCustomKeyHost   `json:"host,omitempty"`
}

type RulesetRuleActionParametersCustomKeyHeader struct {
	RulesetRuleActionParametersCustomKeyFields
	ExcludeOrigin *bool `json:"exclude_origin,omitempty"`
}

type RulesetRuleActionParametersCustomKeyCookie RulesetRuleActionParametersCustomKeyFields

type RulesetRuleActionParametersCustomKeyFields struct {
	Include       []string `json:"include,omitempty"`
	CheckPresence []string `json:"check_presence,omitempty"`
}

type RulesetRuleActionParametersCustomKeyQuery struct {
	Include *RulesetRuleActionParametersCustomKeyList `json:"include,omitempty"`
	Exclude *RulesetRuleActionParametersCustomKeyList `json:"exclude,omitempty"`
	Ignore  *bool                                     `json:"ignore,omitempty"`
}

type RulesetRuleActionParametersCustomKeyList struct {
	List []string
	All  bool
}

func (s *RulesetRuleActionParametersCustomKeyList) UnmarshalJSON(data []byte) error {
	var all string
	if err := json.Unmarshal(data, &all); err == nil {
		s.All = all == "*"
		return nil
	}
	var list []string
	if err := json.Unmarshal(data, &list); err == nil {
		s.List = list
	}

	return nil
}

func (s RulesetRuleActionParametersCustomKeyList) MarshalJSON() ([]byte, error) {
	if s.All {
		return json.Marshal("*")
	}
	return json.Marshal(s.List)
}

type RulesetRuleActionParametersCustomKeyUser struct {
	DeviceType *bool `json:"device_type,omitempty"`
	Geo        *bool `json:"geo,omitempty"`
	Lang       *bool `json:"lang,omitempty"`
}

type RulesetRuleActionParametersCustomKeyHost struct {
	Resolved *bool `json:"resolved,omitempty"`
}

// RulesetRuleActionParametersBlockResponse holds the BlockResponse struct
// for an action parameter.
type RulesetRuleActionParametersBlockResponse struct {
	StatusCode  uint16 `json:"status_code"`
	ContentType string `json:"content_type"`
	Content     string `json:"content"`
}

// RulesetRuleActionParametersURI holds the URI struct for an action parameter.
type RulesetRuleActionParametersURI struct {
	Path   *RulesetRuleActionParametersURIPath  `json:"path,omitempty"`
	Query  *RulesetRuleActionParametersURIQuery `json:"query,omitempty"`
	Origin *bool                                `json:"origin,omitempty"`
}

// RulesetRuleActionParametersURIPath holds the path specific portion of a URI
// action parameter.
type RulesetRuleActionParametersURIPath struct {
	Value      string `json:"value,omitempty"`
	Expression string `json:"expression,omitempty"`
}

// RulesetRuleActionParametersURIQuery holds the query specific portion of a URI
// action parameter.
type RulesetRuleActionParametersURIQuery struct {
	Value      *string `json:"value,omitempty"`
	Expression string  `json:"expression,omitempty"`
}

// RulesetRuleActionParametersHTTPHeader is the definition for define action
// parameters that involve HTTP headers.
type RulesetRuleActionParametersHTTPHeader struct {
	Operation  string `json:"operation,omitempty"`
	Value      string `json:"value,omitempty"`
	Expression string `json:"expression,omitempty"`
}

type RulesetRuleActionParametersOverrides struct {
	Enabled          *bool                                   `json:"enabled,omitempty"`
	Action           string                                  `json:"action,omitempty"`
	SensitivityLevel string                                  `json:"sensitivity_level,omitempty"`
	Categories       []RulesetRuleActionParametersCategories `json:"categories,omitempty"`
	Rules            []RulesetRuleActionParametersRules      `json:"rules,omitempty"`
}

type RulesetRuleActionParametersCategories struct {
	Category string `json:"category"`
	Action   string `json:"action,omitempty"`
	Enabled  *bool  `json:"enabled,omitempty"`
}

type RulesetRuleActionParametersRules struct {
	ID               string `json:"id"`
	Action           string `json:"action,omitempty"`
	Enabled          *bool  `json:"enabled,omitempty"`
	ScoreThreshold   int    `json:"score_threshold,omitempty"`
	SensitivityLevel string `json:"sensitivity_level,omitempty"`
}

// RulesetRuleActionParametersMatchedData holds the structure for WAF based
// payload logging.
type RulesetRuleActionParametersMatchedData struct {
	PublicKey string `json:"public_key,omitempty"`
}

// RulesetRuleActionParametersOrigin is the definition for route action
// parameters that involve origin overrides.
type RulesetRuleActionParametersOrigin struct {
	Host string `json:"host,omitempty"`
	Port uint16 `json:"port,omitempty"`
}

// RulesetRuleActionParametersSni is the definition for the route action
// parameters that involve SNI overrides.
type RulesetRuleActionParametersSni struct {
	Value string `json:"value"`
}

// RulesetRuleActionParametersCompressionAlgorithm defines a compression
// algorithm for the compress_response action.
type RulesetRuleActionParametersCompressionAlgorithm struct {
	Name string `json:"name"`
}

type Polish int

const (
	_ Polish = iota
	PolishOff
	PolishLossless
	PolishLossy
)

func (p Polish) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p Polish) String() string {
	return [...]string{"off", "lossless", "lossy"}[p-1]
}

func (p *Polish) UnmarshalJSON(data []byte) error {
	var (
		s   string
		err error
	)
	err = json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	v, err := PolishFromString(s)
	if err != nil {
		return err
	}
	*p = *v
	return nil
}

func PolishFromString(s string) (*Polish, error) {
	s = strings.ToLower(s)
	var v Polish
	switch s {
	case "off":
		v = PolishOff
	case "lossless":
		v = PolishLossless
	case "lossy":
		v = PolishLossy
	default:
		return nil, fmt.Errorf("unknown variant for polish: %s", s)
	}
	return &v, nil
}

func (p Polish) IntoRef() *Polish {
	return &p
}

type SecurityLevel int

const (
	_ SecurityLevel = iota
	SecurityLevelOff
	SecurityLevelEssentiallyOff
	SecurityLevelLow
	SecurityLevelMedium
	SecurityLevelHigh
	SecurityLevelHelp
)

func (p SecurityLevel) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p SecurityLevel) String() string {
	return [...]string{"off", "essentially_off", "low", "medium", "high", "under_attack"}[p-1]
}

func (p *SecurityLevel) UnmarshalJSON(data []byte) error {
	var (
		s   string
		err error
	)
	err = json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	v, err := SecurityLevelFromString(s)
	if err != nil {
		return err
	}
	*p = *v
	return nil
}

func SecurityLevelFromString(s string) (*SecurityLevel, error) {
	s = strings.ToLower(s)
	var v SecurityLevel
	switch s {
	case "off":
		v = SecurityLevelOff
	case "essentially_off":
		v = SecurityLevelEssentiallyOff
	case "low":
		v = SecurityLevelLow
	case "medium":
		v = SecurityLevelMedium
	case "high":
		v = SecurityLevelHigh
	case "under_attack":
		v = SecurityLevelHelp
	default:
		return nil, fmt.Errorf("unknown variant for security_level: %s", s)
	}
	return &v, nil
}

func (p SecurityLevel) IntoRef() *SecurityLevel {
	return &p
}

type SSL int

const (
	_ SSL = iota
	SSLOff
	SSLFlexible
	SSLFull
	SSLStrict
	SSLOriginPull
)

func (p SSL) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p SSL) String() string {
	return [...]string{"off", "flexible", "full", "strict", "origin_pull"}[p-1]
}

func (p *SSL) UnmarshalJSON(data []byte) error {
	var (
		s   string
		err error
	)
	err = json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	v, err := SSLFromString(s)
	if err != nil {
		return err
	}
	*p = *v
	return nil
}

func SSLFromString(s string) (*SSL, error) {
	s = strings.ToLower(s)
	var v SSL
	switch s {
	case "off":
		v = SSLOff
	case "flexible":
		v = SSLFlexible
	case "full":
		v = SSLFull
	case "strict":
		v = SSLStrict
	case "origin_pull":
		v = SSLOriginPull
	default:
		return nil, fmt.Errorf("unknown variant for ssl: %s", s)
	}
	return &v, nil
}

func (p SSL) IntoRef() *SSL {
	return &p
}

// RulesetRule contains information about a single Ruleset Rule.
type RulesetRule struct {
	ID                     string                             `json:"id,omitempty"`
	Version                *string                            `json:"version,omitempty"`
	Action                 string                             `json:"action"`
	ActionParameters       *RulesetRuleActionParameters       `json:"action_parameters,omitempty"`
	Expression             string                             `json:"expression"`
	Description            string                             `json:"description,omitempty"`
	LastUpdated            *time.Time                         `json:"last_updated,omitempty"`
	Ref                    string                             `json:"ref,omitempty"`
	Enabled                *bool                              `json:"enabled,omitempty"`
	ScoreThreshold         int                                `json:"score_threshold,omitempty"`
	RateLimit              *RulesetRuleRateLimit              `json:"ratelimit,omitempty"`
	ExposedCredentialCheck *RulesetRuleExposedCredentialCheck `json:"exposed_credential_check,omitempty"`
	Logging                *RulesetRuleLogging                `json:"logging,omitempty"`
}

// RulesetRuleRateLimit contains the structure of a HTTP rate limit Ruleset Rule.
type RulesetRuleRateLimit struct {
	Characteristics         []string `json:"characteristics,omitempty"`
	RequestsPerPeriod       int      `json:"requests_per_period,omitempty"`
	ScorePerPeriod          int      `json:"score_per_period,omitempty"`
	ScoreResponseHeaderName string   `json:"score_response_header_name,omitempty"`
	Period                  int      `json:"period,omitempty"`
	MitigationTimeout       int      `json:"mitigation_timeout,omitempty"`
	CountingExpression      string   `json:"counting_expression,omitempty"`
	RequestsToOrigin        bool     `json:"requests_to_origin,omitempty"`
}

// RulesetRuleExposedCredentialCheck contains the structure of an exposed
// credential check Ruleset Rule.
type RulesetRuleExposedCredentialCheck struct {
	UsernameExpression string `json:"username_expression,omitempty"`
	PasswordExpression string `json:"password_expression,omitempty"`
}

// RulesetRuleLogging contains the logging configuration for the rule.
type RulesetRuleLogging struct {
	Enabled *bool `json:"enabled,omitempty"`
}

// UpdateRulesetRequest is the representation of a Ruleset update.
type UpdateRulesetRequest struct {
	Description string        `json:"description"`
	Rules       []RulesetRule `json:"rules"`
}

// ListRulesetResponse contains all Rulesets.
type ListRulesetResponse struct {
	Response
	Result []Ruleset `json:"result"`
}

// GetRulesetResponse contains a single Ruleset.
type GetRulesetResponse struct {
	Response
	Result Ruleset `json:"result"`
}

// CreateRulesetResponse contains response data when creating a new Ruleset.
type CreateRulesetResponse struct {
	Response
	Result Ruleset `json:"result"`
}

// UpdateRulesetResponse contains response data when updating an existing
// Ruleset.
type UpdateRulesetResponse struct {
	Response
	Result Ruleset `json:"result"`
}

type ListRulesetsParams struct{}

type CreateRulesetParams struct {
	Name        string        `json:"name,omitempty"`
	Description string        `json:"description,omitempty"`
	Kind        string        `json:"kind,omitempty"`
	Phase       string        `json:"phase,omitempty"`
	Rules       []RulesetRule `json:"rules"`
}

type UpdateRulesetParams struct {
	ID          string        `json:"-"`
	Description string        `json:"description"`
	Rules       []RulesetRule `json:"rules"`
}

type UpdateEntrypointRulesetParams struct {
	Phase       string        `json:"-"`
	Description string        `json:"description,omitempty"`
	Rules       []RulesetRule `json:"rules"`
}

// ListRulesets lists all Rulesets in a given zone or account depending on the
// ResourceContainer type provided.
//
// API reference: https://developers.cloudflare.com/api/operations/listAccountRulesets
// API reference: https://developers.cloudflare.com/api/operations/listZoneRulesets
func (api *API) ListRulesets(ctx context.Context, rc *ResourceContainer, params ListRulesetsParams) ([]Ruleset, error) {
	uri := fmt.Sprintf("/%s/%s/rulesets", rc.Level, rc.Identifier)

	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []Ruleset{}, err
	}

	result := ListRulesetResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return []Ruleset{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result.Result, nil
}

// GetRuleset fetches a single ruleset.
//
// API reference: https://developers.cloudflare.com/api/operations/getAccountRuleset
// API reference: https://developers.cloudflare.com/api/operations/getZoneRuleset
func (api *API) GetRuleset(ctx context.Context, rc *ResourceContainer, rulesetID string) (Ruleset, error) {
	uri := fmt.Sprintf("/%s/%s/rulesets/%s", rc.Level, rc.Identifier, rulesetID)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return Ruleset{}, err
	}

	result := GetRulesetResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return Ruleset{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result.Result, nil
}

// CreateRuleset creates a new ruleset.
//
// API reference: https://developers.cloudflare.com/api/operations/createAccountRuleset
// API reference: https://developers.cloudflare.com/api/operations/createZoneRuleset
func (api *API) CreateRuleset(ctx context.Context, rc *ResourceContainer, params CreateRulesetParams) (Ruleset, error) {
	uri := fmt.Sprintf("/%s/%s/rulesets", rc.Level, rc.Identifier)
	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, params)
	if err != nil {
		return Ruleset{}, err
	}

	result := CreateRulesetResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return Ruleset{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result.Result, nil
}

// DeleteRuleset removes a ruleset based on the ruleset ID.
//
// API reference: https://developers.cloudflare.com/api/operations/deleteAccountRuleset
// API reference: https://developers.cloudflare.com/api/operations/deleteZoneRuleset
func (api *API) DeleteRuleset(ctx context.Context, rc *ResourceContainer, rulesetID string) error {
	uri := fmt.Sprintf("/%s/%s/rulesets/%s", rc.Level, rc.Identifier, rulesetID)
	res, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}

	// The API is not implementing the standard response blob but returns an
	// empty response (204) in case of a success. So we are checking for the
	// response body size here.
	if len(res) > 0 {
		return fmt.Errorf(errMakeRequestError+": %w", errors.New(string(res)))
	}

	return nil
}

// UpdateRuleset updates a ruleset based on the ruleset ID.
//
// API reference: https://developers.cloudflare.com/api/operations/updateAccountRuleset
// API reference: https://developers.cloudflare.com/api/operations/updateZoneRuleset
func (api *API) UpdateRuleset(ctx context.Context, rc *ResourceContainer, params UpdateRulesetParams) (Ruleset, error) {
	if params.ID == "" {
		return Ruleset{}, ErrMissingResourceIdentifier
	}

	uri := fmt.Sprintf("/%s/%s/rulesets/%s", rc.Level, rc.Identifier, params.ID)
	res, err := api.makeRequestContext(ctx, http.MethodPut, uri, params)
	if err != nil {
		return Ruleset{}, err
	}

	result := UpdateRulesetResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return Ruleset{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result.Result, nil
}

// GetEntrypointRuleset returns an entry point ruleset base on the phase.
//
// API reference: https://developers.cloudflare.com/api/operations/getAccountEntrypointRuleset
// API reference: https://developers.cloudflare.com/api/operations/getZoneEntrypointRuleset
func (api *API) GetEntrypointRuleset(ctx context.Context, rc *ResourceContainer, phase string) (Ruleset, error) {
	uri := fmt.Sprintf("/%s/%s/rulesets/phases/%s/entrypoint", rc.Level, rc.Identifier, phase)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return Ruleset{}, err
	}

	result := GetRulesetResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return Ruleset{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result.Result, nil
}

// UpdateEntrypointRuleset updates an entry point ruleset phase based on the
// phase.
//
// API reference: https://developers.cloudflare.com/api/operations/updateAccountEntrypointRuleset
// API reference: https://developers.cloudflare.com/api/operations/updateZoneEntrypointRuleset
func (api *API) UpdateEntrypointRuleset(ctx context.Context, rc *ResourceContainer, params UpdateEntrypointRulesetParams) (Ruleset, error) {
	if params.Phase == "" {
		return Ruleset{}, ErrMissingRulesetPhase
	}

	uri := fmt.Sprintf("/%s/%s/rulesets/phases/%s/entrypoint", rc.Level, rc.Identifier, params.Phase)
	res, err := api.makeRequestContext(ctx, http.MethodPut, uri, params)
	if err != nil {
		return Ruleset{}, err
	}

	result := GetRulesetResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return Ruleset{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result.Result, nil
}
