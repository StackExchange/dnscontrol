package diff

import (
	"strconv"
	"strings"
	"testing"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/miekg/dns/dnsutil"
)

func myRecord(s string) *models.RecordConfig {
	parts := strings.Split(s, " ")
	ttl, _ := strconv.ParseUint(parts[2], 10, 32)
	return &models.RecordConfig{
		NameFQDN: dnsutil.AddOrigin(parts[0], "example.com"),
		Type:     parts[1],
		TTL:      uint32(ttl),
		Target:   parts[3],
		Metadata: map[string]string{},
	}
}

func TestAdditionsOnly(t *testing.T) {
	desired := []*models.RecordConfig{
		myRecord("@ A 1 1.2.3.4"),
	}
	existing := []*models.RecordConfig{}
	checkLengths(t, existing, desired, 0, 1, 0, 0)
}

func TestDeletionsOnly(t *testing.T) {
	existing := []*models.RecordConfig{
		myRecord("@ A 1 1.2.3.4"),
	}
	desired := []*models.RecordConfig{}
	checkLengths(t, existing, desired, 0, 0, 1, 0)
}

func TestModification(t *testing.T) {
	existing := []*models.RecordConfig{
		myRecord("www A 1 1.1.1.1"),
		myRecord("@ A 1 1.2.3.4"),
	}
	desired := []*models.RecordConfig{
		myRecord("@ A 32 1.2.3.4"),
		myRecord("www A 1 1.1.1.1"),
	}
	un, _, _, mod := checkLengths(t, existing, desired, 1, 0, 0, 1)
	if un[0].Desired != desired[1] || un[0].Existing != existing[0] {
		t.Error("Expected unchanged records to be correlated")
	}
	if mod[0].Desired != desired[0] || mod[0].Existing != existing[1] {
		t.Errorf("Expected modified records to be correlated")
	}
}

func TestUnchangedWithAddition(t *testing.T) {
	existing := []*models.RecordConfig{
		myRecord("www A 1 1.1.1.1"),
	}
	desired := []*models.RecordConfig{
		myRecord("www A 1 1.2.3.4"),
		myRecord("www A 1 1.1.1.1"),
	}
	un, _, _, _ := checkLengths(t, existing, desired, 1, 1, 0, 0)
	if un[0].Desired != desired[1] || un[0].Existing != existing[0] {
		t.Errorf("Expected unchanged records to be correlated")
	}
}

func TestOutOfOrderRecords(t *testing.T) {
	existing := []*models.RecordConfig{
		myRecord("www A 1 1.1.1.1"),
		myRecord("www A 1 2.2.2.2"),
		myRecord("www A 1 3.3.3.3"),
	}
	desired := []*models.RecordConfig{
		myRecord("www A 1 1.1.1.1"),
		myRecord("www A 1 2.2.2.2"),
		myRecord("www A 1 2.2.2.3"),
		myRecord("www A 10 3.3.3.3"),
	}
	_, _, _, mods := checkLengths(t, existing, desired, 2, 1, 0, 1)
	if mods[0].Desired != desired[3] || mods[0].Existing != existing[2] {
		t.Fatalf("Expected to match %s and %s, but matched %s and %s", existing[2], desired[3], mods[0].Existing, mods[0].Desired)
	}
}

func TestMxPrio(t *testing.T) {
	existing := []*models.RecordConfig{
		myRecord("www MX 1 1.1.1.1"),
	}
	desired := []*models.RecordConfig{
		myRecord("www MX 1 1.1.1.1"),
	}
	existing[0].MxPreference = 10
	desired[0].MxPreference = 20
	checkLengths(t, existing, desired, 0, 0, 0, 1)
}

func TestTTLChange(t *testing.T) {
	existing := []*models.RecordConfig{
		myRecord("www MX 1 1.1.1.1"),
	}
	desired := []*models.RecordConfig{
		myRecord("www MX 10 1.1.1.1"),
	}
	checkLengths(t, existing, desired, 0, 0, 0, 1)
}

func TestMetaChange(t *testing.T) {
	existing := []*models.RecordConfig{
		myRecord("www MX 1 1.1.1.1"),
	}
	desired := []*models.RecordConfig{
		myRecord("www MX 1 1.1.1.1"),
	}
	existing[0].Metadata["k"] = "aa"
	desired[0].Metadata["k"] = "bb"
	checkLengths(t, existing, desired, 1, 0, 0, 0)
	getMeta := func(r *models.RecordConfig) map[string]string {
		return map[string]string{
			"k": r.Metadata["k"],
		}
	}
	checkLengths(t, existing, desired, 0, 0, 0, 1, getMeta)
}

func checkLengths(t *testing.T, existing, desired []*models.RecordConfig, unCount, createCount, delCount, modCount int, valFuncs ...func(*models.RecordConfig) map[string]string) (un, cre, del, mod Changeset) {
	return checkLengthsFull(t, existing, desired, unCount, createCount, delCount, modCount, false, valFuncs...)
}

func checkLengthsFull(t *testing.T, existing, desired []*models.RecordConfig, unCount, createCount, delCount, modCount int, keepUnknown bool, valFuncs ...func(*models.RecordConfig) map[string]string) (un, cre, del, mod Changeset) {
	dc := &models.DomainConfig{
		Name:        "example.com",
		Records:     desired,
		KeepUnknown: keepUnknown,
	}
	d := New(dc, valFuncs...)
	un, cre, del, mod = d.IncrementalDiff(existing)
	if len(un) != unCount {
		t.Errorf("Got %d unchanged records, but expected %d", len(un), unCount)
	}
	if len(cre) != createCount {
		t.Errorf("Got %d records to create, but expected %d", len(cre), createCount)
	}
	if len(del) != delCount {
		t.Errorf("Got %d records to delete, but expected %d", len(del), delCount)
	}
	if len(mod) != modCount {
		t.Errorf("Got %d records to modify, but expected %d", len(mod), modCount)
	}
	if t.Failed() {
		t.FailNow()
	}
	return
}

func TestNoPurge(t *testing.T) {
	existing := []*models.RecordConfig{
		myRecord("www MX 1 1.1.1.1"),
		myRecord("www MX 1 2.2.2.2"),
		myRecord("www2 MX 1 1.1.1.1"),
	}
	desired := []*models.RecordConfig{
		myRecord("www MX 1 1.1.1.1"),
	}
	checkLengthsFull(t, existing, desired, 1, 0, 1, 0, true)
}
