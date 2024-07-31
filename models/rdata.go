package models

// Rdataer is an interface for resource types.
type Rdataer interface {
	Name() string                  // Return the rtype name used in RecordConfig.Name ("MX", etc.)
	ComputeTarget() string         // Pre-compute the value stored in RecordConfig.target.
	ComputeComparableMini() string // Pre-compute the value stored in RecordConfig.ComparableMini.
}

// Seal finalizes a RecordConfig by setting .Rdata and pre-computing
// various values.
func (rc *RecordConfig) Seal(zone string, shortLabel string, rdata Rdataer) {
	rc.Type = rdata.Name()
	rc.SetLabel(shortLabel, zone)
	rc.Rdata = rdata

	rc.ReSeal() // Fill in the pre-computed fields.
}

// ReSeal re-computes the fields that are pre-computed.
func (rc *RecordConfig) ReSeal() {
	rdata := rc.Rdata
	rc.SetTarget(rdata.ComputeTarget())
	rc.ComparableMini = rc.Rdata.ComputeComparableMini()
}
