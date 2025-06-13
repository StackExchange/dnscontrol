package none

// EnsureZoneExists ensures that the zone exists on the DNS server.  It will create it if it does not.
// This is a no-op for the None provider; it pretends that all zones already exist.
func (n None) EnsureZoneExists(domain string) error {
	return nil
}
