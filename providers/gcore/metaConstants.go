package gcore

const (
	// Applies to RRSet
	metaFilters = "gcore_filters"

	// Only applies to RRSet metadata
	metaFailoverProtocol       = "gcore_failover_protocol"
	metaFailoverPort           = "gcore_failover_port"
	metaFailoverFrequency      = "gcore_failover_frequency"
	metaFailoverTimeout        = "gcore_failover_timeout"
	metaFailoverMethod         = "gcore_failover_method"
	metaFailoverCommand        = "gcore_failover_command"
	metaFailoverURL            = "gcore_failover_url"
	metaFailoverTLS            = "gcore_failover_tls"
	metaFailoverRegexp         = "gcore_failover_regexp"
	metaFailoverHTTPStatusCode = "gcore_failover_http_status_code"
	metaFailoverHost           = "gcore_failover_host"

	// Only applies to record metadata
	metaASN        = "gcore_asn"
	metaContinents = "gcore_continents"
	metaCountries  = "gcore_countries"
	metaLatitude   = "gcore_latitude"
	metaLongitude  = "gcore_longitude"
	metaFallback   = "gcore_fallback"
	metaBackup     = "gcore_backup"
	metaNotes      = "gcore_notes"
	metaWeight     = "gcore_weight"
	metaIP         = "gcore_ip"
)
