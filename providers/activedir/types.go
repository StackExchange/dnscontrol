package activedir

// DNSAccessor describes a system that can access Microsoft DNS.
type DNSAccessor interface {
	Exit()
	GetDNSServerZoneAll() ([]string, error)
}
