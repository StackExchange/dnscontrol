package diff2

import (
	"reflect"
	"testing"

	"github.com/StackExchange/dnscontrol/v3/models"
)

func makeRec(label, rtype, content string) *models.RecordConfig {
	origin := "f.com"
	r := models.RecordConfig{}
	r.SetLabel(label, origin)
	r.PopulateFromString(rtype, content, origin)
	return &r
}
func makeRecSet(recs ...*models.RecordConfig) *recset {
	result := recset{}
	result.Key = recs[0].Key()
	result.Recs = append(result.Recs, recs...)
	return &result
}
func Test_groupbyRSet(t *testing.T) {

	wwwa1 := makeRec("www", "A", "1.1.1.1")
	wwwa2 := makeRec("www", "A", "2.2.2.2")
	zzza1 := makeRec("zzz", "A", "1.1.0.0")
	zzza2 := makeRec("zzz", "A", "2.2.0.0")
	wwwmx1 := makeRec("www", "MX", "1 mx1.foo.com.")
	wwwmx2 := makeRec("www", "MX", "2 mx2.foo.com.")
	zzzmx1 := makeRec("zzz", "MX", "1 mx.foo.com.")
	orig := models.Records{wwwa1, wwwa2, zzza1, zzza2, wwwmx1, wwwmx2, zzzmx1}
	wantResult := []recset{
		*makeRecSet(wwwa1, wwwa2),
		*makeRecSet(wwwmx1, wwwmx2),
		*makeRecSet(zzza1, zzza2),
		*makeRecSet(zzzmx1),
	}

	t.Run("afew", func(t *testing.T) {
		if gotResult := groupbyRSet(orig, "f.com"); !reflect.DeepEqual(gotResult, wantResult) {
			t.Errorf("groupbyRSet() = %v, want %v", gotResult, wantResult)
		}
	})
}
