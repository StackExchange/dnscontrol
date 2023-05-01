package bindserial

// NB(tlim): Yes, its gross to use a global variable for this.
// However there's no cleaner way to do it.  Ideally we'd add a way to
// have per-provider flags or settings on the command line.  At least
// by isolating it to this file we limit the blast radius of this bad
// decision.

// ForcedValue if non-zero, BIND will generate SOA serial numbers
// using this value.
var ForcedValue int64
