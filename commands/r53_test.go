package commands

import (
	"testing"

	"github.com/StackExchange/dnscontrol/v3/models"
	_ "github.com/StackExchange/dnscontrol/v3/providers/_all"
)

func TestR53Test_1(t *testing.T) {
	rec := models.RecordConfig{
		Type:     "R53_ALIAS",
		Name:     "foo",
		NameFQDN: "foo.domain.tld",
		Target:   "bar",
	}
	rec.R53Alias = make(map[string]string)
	rec.R53Alias["type"] = "A"
	w := `R53_ALIAS('foo', 'A', 'bar')`
	if g := makeR53alias(&rec, 0); g != w {
		t.Errorf("makeR53alias failure: got `%s` want `%s`", g, w)
	}
}

func TestR53Test_1ttl(t *testing.T) {
	rec := models.RecordConfig{
		Type:     "R53_ALIAS",
		Name:     "foo",
		NameFQDN: "foo.domain.tld",
		Target:   "bar",
	}
	rec.R53Alias = make(map[string]string)
	rec.R53Alias["type"] = "A"
	w := `R53_ALIAS('foo', 'A', 'bar', TTL(321))`
	if g := makeR53alias(&rec, 321); g != w {
		t.Errorf("makeR53alias failure: got `%s` want `%s`", g, w)
	}
}

func TestR53Test_2(t *testing.T) {
	rec := models.RecordConfig{
		Type:     "R53_ALIAS",
		Name:     "foo",
		NameFQDN: "foo.domain.tld",
		Target:   "bar",
	}
	rec.R53Alias = make(map[string]string)
	rec.R53Alias["type"] = "A"
	rec.R53Alias["zone_id"] = "blarg"
	w := `R53_ALIAS('foo', 'A', 'bar', R53_ZONE('blarg'))`
	if g := makeR53alias(&rec, 0); g != w {
		t.Errorf("makeR53alias failure: got `%s` want `%s`", g, w)
	}
}

func TestR53Test_2ttl(t *testing.T) {
	rec := models.RecordConfig{
		Type:     "R53_ALIAS",
		Name:     "foo",
		NameFQDN: "foo.domain.tld",
		Target:   "bar",
	}
	rec.R53Alias = make(map[string]string)
	rec.R53Alias["type"] = "A"
	rec.R53Alias["zone_id"] = "blarg"
	w := `R53_ALIAS('foo', 'A', 'bar', R53_ZONE('blarg'), TTL(123))`
	if g := makeR53alias(&rec, 123); g != w {
		t.Errorf("makeR53alias failure: got `%s` want `%s`", g, w)
	}
}
