package models

type Rdataer interface {
	ComputeTarget() string
}

func (rc *RecordConfig) Seal(typeName string, zone string, shortLabel string, rdata Rdataer) {

	//TODO: Verify that rc.Rdata is of type rc.Type.

	rc.Type = typeName
	rc.SetLabel(shortLabel, zone)
	rc.Rdata = rdata
	rc.SetTarget(rdata.ComputeTarget())
	//rc.SetComparable(rc.Rdata.ComputeComparable())

}
