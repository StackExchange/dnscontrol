package gcloud

import (
	"time"

	"github.com/StackExchange/dnscontrol/v4/models"
	gdns "google.golang.org/api/dns/v1"
)

func (g *gcloudProvider) getDnssecCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	enabled, err := g.isDnssecEnabled(dc.Name)
	if err != nil {
		return nil, err
	}
	if enabled && dc.AutoDNSSEC == "off" {
		return []*models.Correction{
			{
				Msg: "Disable DNSSEC",
				F:   func() error { err := g.disableDnssec(dc.Name); return err },
			},
		}, nil
	}

	if !enabled && dc.AutoDNSSEC == "on" {
		return []*models.Correction{
			{
				Msg: "Enable DNSSEC",
				F:   func() error { err := g.enableDnssec(dc.Name); return err },
			},
		}, nil
	}

	return nil, nil
}

func (g *gcloudProvider) isDnssecEnabled(domain string) (bool, error) {
	// Zones that never had DNSSEC enabled will have nil in the DnssecConfig field
	if g.zones[domain+"."].DnssecConfig == nil {
		return false, nil
	}
	if g.zones[domain+"."].DnssecConfig.State == "on" {
		return true, nil
	}
	// Google Cloud DNS has a "transfer" state for DNSSEC. We treat it as "on".
	// It basically means that DNSSEC is enabled, but Google won't rotate ZSKs.
	if g.zones[domain+"."].DnssecConfig.State == "transfer" {
		return true, nil
	}
	return false, nil
}

func (g *gcloudProvider) enableDnssec(domain string) error {
	dnssecPatch := &gdns.ManagedZone{
		DnssecConfig: &gdns.ManagedZoneDnsSecConfig{
			State: "on",
		},
	}
	resp, err := g.client.ManagedZones.Patch(g.project, g.zones[domain+"."].Name, dnssecPatch).Do()
	if err != nil {
		return err
	}
	if resp.Status != "done" {
		// Should we have a timeout here?
		for {
			checkOperation, err := g.client.ManagedZoneOperations.Get(g.project, g.zones[domain+"."].Name, resp.Id).Do()
			if err != nil {
				return err
			}
			if checkOperation.Status == "done" {
				return nil
			}
			time.Sleep(2 * time.Second)
		}
	}

	return nil
}

func (g *gcloudProvider) disableDnssec(domain string) error {
	dnssecPatch := &gdns.ManagedZone{
		DnssecConfig: &gdns.ManagedZoneDnsSecConfig{
			State: "off",
		},
	}
	resp, err := g.client.ManagedZones.Patch(g.project, g.zones[domain+"."].Name, dnssecPatch).Do()
	if err != nil {
		return err
	}
	if resp.Status != "done" {
		// Should we have a timeout here?
		for {
			checkOperation, err := g.client.ManagedZoneOperations.Get(g.project, g.zones[domain+"."].Name, resp.Id).Do()
			if err != nil {
				return err
			}
			if checkOperation.Status == "done" {
				return nil
			}
			time.Sleep(2 * time.Second)
		}
	}

	return nil
}
