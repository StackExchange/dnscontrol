package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

var ErrMissingClusterID = errors.New("missing required cluster ID")

// DNSFirewallCluster represents a DNS Firewall configuration.
type DNSFirewallCluster struct {
	ID                   string   `json:"id,omitempty"`
	Name                 string   `json:"name"`
	UpstreamIPs          []string `json:"upstream_ips"`
	DNSFirewallIPs       []string `json:"dns_firewall_ips,omitempty"`
	MinimumCacheTTL      uint     `json:"minimum_cache_ttl,omitempty"`
	MaximumCacheTTL      uint     `json:"maximum_cache_ttl,omitempty"`
	DeprecateAnyRequests bool     `json:"deprecate_any_requests"`
	ModifiedOn           string   `json:"modified_on,omitempty"`
}

// DNSFirewallAnalyticsMetrics represents a group of aggregated DNS Firewall metrics.
type DNSFirewallAnalyticsMetrics struct {
	QueryCount         *int64   `json:"queryCount"`
	UncachedCount      *int64   `json:"uncachedCount"`
	StaleCount         *int64   `json:"staleCount"`
	ResponseTimeAvg    *float64 `json:"responseTimeAvg"`
	ResponseTimeMedian *float64 `json:"responseTimeMedian"`
	ResponseTime90th   *float64 `json:"responseTime90th"`
	ResponseTime99th   *float64 `json:"responseTime99th"`
}

// DNSFirewallAnalytics represents a set of aggregated DNS Firewall metrics.
// TODO: Add the queried data and not only the aggregated values.
type DNSFirewallAnalytics struct {
	Totals DNSFirewallAnalyticsMetrics `json:"totals"`
	Min    DNSFirewallAnalyticsMetrics `json:"min"`
	Max    DNSFirewallAnalyticsMetrics `json:"max"`
}

// DNSFirewallUserAnalyticsOptions represents range and dimension selection on analytics endpoint.
type DNSFirewallUserAnalyticsOptions struct {
	Metrics []string   `url:"metrics,omitempty" del:","`
	Since   *time.Time `url:"since,omitempty"`
	Until   *time.Time `url:"until,omitempty"`
}

// dnsFirewallResponse represents a DNS Firewall response.
type dnsFirewallResponse struct {
	Response
	Result *DNSFirewallCluster `json:"result"`
}

// dnsFirewallListResponse represents an array of DNS Firewall responses.
type dnsFirewallListResponse struct {
	Response
	Result []*DNSFirewallCluster `json:"result"`
}

// dnsFirewallAnalyticsResponse represents a DNS Firewall analytics response.
type dnsFirewallAnalyticsResponse struct {
	Response
	Result DNSFirewallAnalytics `json:"result"`
}

type CreateDNSFirewallClusterParams struct {
	Name                 string   `json:"name"`
	UpstreamIPs          []string `json:"upstream_ips"`
	DNSFirewallIPs       []string `json:"dns_firewall_ips,omitempty"`
	MinimumCacheTTL      uint     `json:"minimum_cache_ttl,omitempty"`
	MaximumCacheTTL      uint     `json:"maximum_cache_ttl,omitempty"`
	DeprecateAnyRequests bool     `json:"deprecate_any_requests"`
}

type GetDNSFirewallClusterParams struct {
	ClusterID string `json:"-"`
}

type UpdateDNSFirewallClusterParams struct {
	ClusterID            string   `json:"-"`
	Name                 string   `json:"name"`
	UpstreamIPs          []string `json:"upstream_ips"`
	DNSFirewallIPs       []string `json:"dns_firewall_ips,omitempty"`
	MinimumCacheTTL      uint     `json:"minimum_cache_ttl,omitempty"`
	MaximumCacheTTL      uint     `json:"maximum_cache_ttl,omitempty"`
	DeprecateAnyRequests bool     `json:"deprecate_any_requests"`
}

type ListDNSFirewallClustersParams struct{}

type GetDNSFirewallUserAnalyticsParams struct {
	ClusterID string `json:"-"`
	DNSFirewallUserAnalyticsOptions
}

// CreateDNSFirewallCluster creates a new DNS Firewall cluster.
//
// API reference: https://api.cloudflare.com/#dns-firewall-create-dns-firewall-cluster
func (api *API) CreateDNSFirewallCluster(ctx context.Context, rc *ResourceContainer, params CreateDNSFirewallClusterParams) (*DNSFirewallCluster, error) {
	uri := fmt.Sprintf("/%s/dns_firewall", rc.URLFragment())
	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, params)
	if err != nil {
		return nil, err
	}

	response := &dnsFirewallResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return response.Result, nil
}

// GetDNSFirewallCluster fetches a single DNS Firewall cluster.
//
// API reference: https://api.cloudflare.com/#dns-firewall-dns-firewall-cluster-details
func (api *API) GetDNSFirewallCluster(ctx context.Context, rc *ResourceContainer, params GetDNSFirewallClusterParams) (*DNSFirewallCluster, error) {
	if params.ClusterID == "" {
		return &DNSFirewallCluster{}, ErrMissingClusterID
	}

	uri := fmt.Sprintf("/%s/dns_firewall/%s", rc.URLFragment(), params.ClusterID)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	response := &dnsFirewallResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return response.Result, nil
}

// ListDNSFirewallClusters lists the DNS Firewall clusters associated with an account.
//
// API reference: https://api.cloudflare.com/#dns-firewall-list-dns-firewall-clusters
func (api *API) ListDNSFirewallClusters(ctx context.Context, rc *ResourceContainer, params ListDNSFirewallClustersParams) ([]*DNSFirewallCluster, error) {
	uri := fmt.Sprintf("/%s/dns_firewall", rc.URLFragment())
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	response := &dnsFirewallListResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return response.Result, nil
}

// UpdateDNSFirewallCluster updates a DNS Firewall cluster.
//
// API reference: https://api.cloudflare.com/#dns-firewall-update-dns-firewall-cluster
func (api *API) UpdateDNSFirewallCluster(ctx context.Context, rc *ResourceContainer, params UpdateDNSFirewallClusterParams) error {
	if params.ClusterID == "" {
		return ErrMissingClusterID
	}

	uri := fmt.Sprintf("/%s/dns_firewall/%s", rc.URLFragment(), params.ClusterID)
	res, err := api.makeRequestContext(ctx, http.MethodPatch, uri, params)
	if err != nil {
		return err
	}

	response := &dnsFirewallResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return nil
}

// DeleteDNSFirewallCluster deletes a DNS Firewall cluster. Note that this cannot be
// undone, and will stop all traffic to that cluster.
//
// API reference: https://api.cloudflare.com/#dns-firewall-delete-dns-firewall-cluster
func (api *API) DeleteDNSFirewallCluster(ctx context.Context, rc *ResourceContainer, clusterID string) error {
	uri := fmt.Sprintf("/%s/dns_firewall/%s", rc.URLFragment(), clusterID)
	res, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}

	response := &dnsFirewallResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return nil
}

// GetDNSFirewallUserAnalytics retrieves analytics report for a specified dimension and time range.
func (api *API) GetDNSFirewallUserAnalytics(ctx context.Context, rc *ResourceContainer, params GetDNSFirewallUserAnalyticsParams) (DNSFirewallAnalytics, error) {
	uri := buildURI(fmt.Sprintf("/%s/dns_firewall/%s/dns_analytics/report", rc.URLFragment(), params.ClusterID), params.DNSFirewallUserAnalyticsOptions)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return DNSFirewallAnalytics{}, err
	}

	response := dnsFirewallAnalyticsResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return DNSFirewallAnalytics{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return response.Result, nil
}
