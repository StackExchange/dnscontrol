package models

type Rdataer interface {
	Name() string
	ComputeTarget() string
}

func (rc *RecordConfig) Seal(zone string, shortLabel string, rdata Rdataer) {

	rc.Type = rdata.Name()
	rc.SetLabel(shortLabel, zone)
	rc.Rdata = rdata
	rc.SetTarget(rdata.ComputeTarget())
	//rc.SetComparable(rc.Rdata.ComputeComparable())

}
