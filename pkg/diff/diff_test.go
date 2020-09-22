package diff

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/StackExchange/dnscontrol/v3/models"
)

func myRecord(s string) *models.RecordConfig {
	parts := strings.Split(s, " ")
	ttl, _ := strconv.ParseUint(parts[2], 10, 32)
	r := &models.RecordConfig{
		Type:     parts[1],
		TTL:      uint32(ttl),
		Metadata: map[string]string{},
	}
	r.SetLabel(parts[0], "example.com")
	r.SetTarget(parts[3])
	return r
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

// s stringifies a RecordConfig for testing purposes.
func s(rc *models.RecordConfig) string {
	return fmt.Sprintf("%s %s %d %s", rc.GetLabel(), rc.Type, rc.TTL, rc.GetTargetCombined())
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
	if s(mods[0].Desired) != s(desired[3]) || s(mods[0].Existing) != s(existing[2]) {
		t.Fatalf("Expected to match %s and %s, but matched %s and %s", s(existing[2]), s(desired[3]), s(mods[0].Existing), s(mods[0].Desired))
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

func TestMetaOrdering(t *testing.T) {
	existing := []*models.RecordConfig{
		myRecord("www MX 1 1.1.1.1"),
	}
	desired := []*models.RecordConfig{
		myRecord("www MX 1 1.1.1.1"),
	}
	existing[0].Metadata["k"] = "aa"
	existing[0].Metadata["x"] = "cc"
	desired[0].Metadata["k"] = "aa"
	desired[0].Metadata["x"] = "cc"
	checkLengths(t, existing, desired, 1, 0, 0, 0)
	getMeta := func(r *models.RecordConfig) map[string]string {
		return map[string]string{
			"k": r.Metadata["k"],
		}
	}
	checkLengths(t, existing, desired, 1, 0, 0, 0, getMeta)
}

func checkLengths(t *testing.T, existing, desired []*models.RecordConfig, unCount, createCount, delCount, modCount int, valFuncs ...func(*models.RecordConfig) map[string]string) (un, cre, del, mod Changeset) {
	return checkLengthsWithKeepUnknown(t, existing, desired, unCount, createCount, delCount, modCount, false, valFuncs...)
}

func checkLengthsWithKeepUnknown(t *testing.T, existing, desired []*models.RecordConfig, unCount, createCount, delCount, modCount int, keepUnknown bool, valFuncs ...func(*models.RecordConfig) map[string]string) (un, cre, del, mod Changeset) {
	return checkLengthsFull(t, existing, desired, unCount, createCount, delCount, modCount, keepUnknown, []string{}, nil, valFuncs...)
}

func checkLengthsFull(t *testing.T, existing, desired []*models.RecordConfig, unCount, createCount, delCount, modCount int, keepUnknown bool, ignoredRecords []string, ignoredTargets []*models.IgnoreTarget, valFuncs ...func(*models.RecordConfig) map[string]string) (un, cre, del, mod Changeset) {
	dc := &models.DomainConfig{
		Name:           "example.com",
		Records:        desired,
		KeepUnknown:    keepUnknown,
		IgnoredNames:   ignoredRecords,
		IgnoredTargets: ignoredTargets,
	}
	d := New(dc, valFuncs...)
	un, cre, del, mod, err := d.IncrementalDiff(existing)
	if err != nil {
		panic(err)
	}
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
	checkLengthsWithKeepUnknown(t, existing, desired, 1, 0, 1, 0, true)
}

func TestIgnoredRecords(t *testing.T) {
	existing := []*models.RecordConfig{
		myRecord("www1 MX 1 1.1.1.1"),
		myRecord("www2 MX 1 1.1.1.1"),
		myRecord("www3 MX 1 1.1.1.1"),
	}
	desired := []*models.RecordConfig{
		myRecord("www3 MX 1 2.2.2.2"),
	}
	checkLengthsFull(t, existing, desired, 0, 0, 0, 1, false, []string{"www1", "www2"}, nil)
}

func TestModifyingIgnoredRecords(t *testing.T) {
	existing := []*models.RecordConfig{
		myRecord("www1 MX 1 1.1.1.1"),
		myRecord("www2 MX 1 1.1.1.1"),
		myRecord("www3 MX 1 1.1.1.1"),
	}
	desired := []*models.RecordConfig{
		myRecord("www2 MX 1 2.2.2.2"),
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("should panic: modification of IGNOREd record")
		}
	}()

	checkLengthsFull(t, existing, desired, 0, 0, 0, 1, false, []string{"www1", "www2"}, nil)
}

func TestGlobIgnoredName(t *testing.T) {
	existing := []*models.RecordConfig{
		myRecord("www1 MX 1 1.1.1.1"),
		myRecord("foo.www2 MX 1 1.1.1.1"),
		myRecord("foo.bar.www3 MX 1 1.1.1.1"),
		myRecord("www4 MX 1 1.1.1.1"),
	}
	desired := []*models.RecordConfig{
		myRecord("www4 MX 1 2.2.2.2"),
	}
	checkLengthsFull(t, existing, desired, 0, 0, 0, 1, false, []string{"www1", "*.www2", "**.www3"}, nil)
}

func TestInvalidGlobIgnoredName(t *testing.T) {
	existing := []*models.RecordConfig{
		myRecord("www1 MX 1 1.1.1.1"),
		myRecord("www2 MX 1 1.1.1.1"),
		myRecord("www3 MX 1 1.1.1.1"),
	}
	desired := []*models.RecordConfig{
		myRecord("www4 MX 1 2.2.2.2"),
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("should panic: invalid glob pattern for IGNORE_NAME")
		}
	}()

	checkLengthsFull(t, existing, desired, 0, 1, 0, 0, false, []string{"www1", "www2", "[.www3"}, nil)
}

func TestGlobIgnoredTarget(t *testing.T) {
	existing := []*models.RecordConfig{
		myRecord("www1 CNAME 1 ignoreme.com"),
		myRecord("foo.www2 MX 1 1.1.1.2"),
		myRecord("foo.bar.www3 MX 1 1.1.1.1"),
		myRecord("www4 MX 1 1.1.1.1"),
	}
	desired := []*models.RecordConfig{
		myRecord("foo.www2 MX 1 1.1.1.2"),
		myRecord("foo.bar.www3 MX 1 1.1.1.1"),
		myRecord("www4 MX 1 2.2.2.2"),
	}
	checkLengthsFull(t, existing, desired, 2, 0, 0, 1, false, nil, []*models.IgnoreTarget{{Pattern: "ignoreme.com", Type: "CNAME"}})
}

func TestInvalidGlobIgnoredTarget(t *testing.T) {
	existing := []*models.RecordConfig{
		myRecord("www1 MX 1 1.1.1.1"),
		myRecord("www2 MX 1 1.1.1.1"),
		myRecord("www3 MX 1 1.1.1.1"),
	}
	desired := []*models.RecordConfig{
		myRecord("www4 MX 1 2.2.2.2"),
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("should panic: invalid glob pattern for IGNORE_TARGET")
		}
	}()

	checkLengthsFull(t, existing, desired, 0, 1, 0, 0, false, nil, []*models.IgnoreTarget{{Pattern: "[.www3", Type: "CNAME"}})
}

func TestInvalidTypeIgnoredTarget(t *testing.T) {
	existing := []*models.RecordConfig{
		myRecord("www1 MX 1 1.1.1.1"),
		myRecord("www2 MX 1 1.1.1.1"),
		myRecord("www3 MX 1 1.1.1.1"),
	}
	desired := []*models.RecordConfig{
		myRecord("www4 MX 1 2.2.2.2"),
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("should panic: Invalid rType for IGNORE_TARGET A")
		}
	}()

	checkLengthsFull(t, existing, desired, 0, 1, 0, 0, false, nil, []*models.IgnoreTarget{{Pattern: "1.1.1.1", Type: "A"}})
}

// from https://github.com/StackExchange/dnscontrol/issues/552
func TestCaas(t *testing.T) {
	existing := []*models.RecordConfig{
		myRecord("test CAA 1 1.1.1.1"),
		myRecord("test CAA 1 1.1.1.1"),
		myRecord("test CAA 1 1.1.1.1"),
	}
	desired := []*models.RecordConfig{
		myRecord("test CAA 1 1.1.1.1"),
		myRecord("test CAA 1 1.1.1.1"),
		myRecord("test CAA 1 1.1.1.1"),
	}
	existing[0].SetTargetCAA(3, "issue", "letsencrypt.org.")
	existing[1].SetTargetCAA(3, "issue", "amazon.com.")
	existing[2].SetTargetCAA(3, "issuewild", "letsencrypt.org.")

	// this will pass or fail depending on the ordering. Not ok.
	desired[0].SetTargetCAA(3, "issue", "letsencrypt.org.")
	desired[1].SetTargetCAA(3, "issue", "amazon.com.")
	desired[2].SetTargetCAA(3, "issuewild", "letsencrypt.org.")

	checkLengthsFull(t, existing, desired, 3, 0, 0, 0, false, nil, nil)

	// Make sure it passes with a different ordering. Not ok.
	desired[2].SetTargetCAA(3, "issue", "letsencrypt.org.")
	desired[1].SetTargetCAA(3, "issue", "amazon.com.")
	desired[0].SetTargetCAA(3, "issuewild", "letsencrypt.org.")

	checkLengthsFull(t, existing, desired, 3, 0, 0, 0, false, nil, nil)
}
