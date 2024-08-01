package models

// Rdataer is an interface for resource types.
type Rdataer interface {
	// Return the rtype name used in RecordConfig.Name ("MX", etc.)
	Name() string

	// Pre-compute the value stored in RecordConfig.target.
	ComputeTarget() string

	// Pre-compute the value stored in RecordConfig.ComparableMini.
	ComputeComparableMini() string
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
	if rc.Rdata == nil {
		panic("Uninitialized Rdata")
	}
	rc.SetTarget(rc.Rdata.ComputeTarget())
	rc.ComparableMini = rc.Rdata.ComputeComparableMini()
}
