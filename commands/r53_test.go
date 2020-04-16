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
	if g := makeR53alias(&rec, ""); g != w {
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
	if g := makeR53alias(&rec, ""); g != w {
		t.Errorf("makeR53alias failure: got `%s` want `%s`", g, w)
	}
}
