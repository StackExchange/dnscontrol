package providers

//go:generate stringer -type SupportLevel -trimprefix SupportLevel

type SupportLevel int

const (
	SupportLevelDeprecated SupportLevel = iota
	SupportLevelNeedsVolunteer
	SupportLevelCommunity
	SupportLevelOfficial
)
