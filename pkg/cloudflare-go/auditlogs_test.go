package cloudflare

import (
	"strings"
	"testing"
)

func TestAuditLogFilterToQuery(t *testing.T) {
	filter := AuditLogFilter{
		ID: "aaaa",
	}
	if !strings.Contains(filter.ToQuery().Encode(), "id=aaaa") {
		t.Fatalf("Did not properly stringify the id field: %s", filter.ToQuery().Encode())
	}

	filter.ActorEmail = "admin@example.com"
	if !strings.Contains(filter.ToQuery().Encode(), "actor.email=admin%40example.com") {
		t.Fatalf("Did not properly stringify the actor.email field: %s", filter.ToQuery().Encode())
	}

	filter.HideUserLogs = true
	if !strings.Contains(filter.ToQuery().Encode(), "hide_user_logs=true") {
		t.Fatalf("Did not properly stringify the hide_user_logs field: %s", filter.ToQuery().Encode())
	}

	filter.ActorIP = "192.0.2.0"
	if !strings.Contains(filter.ToQuery().Encode(), "&actor.ip=192.0.2.0") {
		t.Fatalf("Did not properly stringify the actorip field: %s", filter.ToQuery().Encode())
	}

	filter.ZoneName = "example.com"
	if !strings.Contains(filter.ToQuery().Encode(), "&zone.name=example.com") {
		t.Fatalf("Did not properly stringify the zone.name field: %s", filter.ToQuery().Encode())
	}

	filter.Direction = "direction"
	if !strings.Contains(filter.ToQuery().Encode(), "&direction=direction") {
		t.Fatalf("Did not properly stringify the direction field: %s", filter.ToQuery().Encode())
	}

	filter.Since = "10-2-2018"
	if !strings.Contains(filter.ToQuery().Encode(), "&since=10-2-2018") {
		t.Fatalf("Did not properly stringify the since field: %s", filter.ToQuery().Encode())
	}

	filter.Before = "10-2-2018"
	if !strings.Contains(filter.ToQuery().Encode(), "&before=10-2-2018") {
		t.Fatalf("Did not properly stringify the before field: %s", filter.ToQuery().Encode())
	}

	filter.PerPage = 10000
	if !strings.Contains(filter.ToQuery().Encode(), "&per_page=10000") {
		t.Fatalf("Did not properly stringify the per_page field: %s", filter.ToQuery().Encode())
	}

	filter.Page = 3
	if !strings.Contains(filter.ToQuery().Encode(), "&page=3") {
		t.Fatalf("Did not properly stringify the page field: %s", filter.ToQuery().Encode())
	}
}
