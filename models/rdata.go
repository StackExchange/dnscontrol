package models

type Rdataer interface {
	Name() string                  // Return the rtype name used in RecordConfig.Name ()"MX", etc.)
	ComputeTarget() string         // Compute the value stored in RecordConfig.target.
	ComputeComparableMini() string // Compute the value stored in RecordConfig.ComparableMini.
}

func (rc *RecordConfig) Seal(zone string, shortLabel string, rdata Rdataer) {

	rc.Type = rdata.Name()
	rc.SetLabel(shortLabel, zone)
	rc.Rdata = rdata
	rc.SetTarget(rdata.ComputeTarget())
	rc.ComparableMini = rc.Rdata.ComputeComparableMini()
}
