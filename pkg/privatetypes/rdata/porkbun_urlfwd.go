package privatetypesrdata

type PORKBUN_URLFWD struct {
	// Deprecated.  Leaving this empty for now. I think the provider substiutes the replacement (URL or URL301) in the provider code, so we don't need to do anything here.
}

func (rd PORKBUN_URLFWD) Len() int {
	return 0
}

func (rd PORKBUN_URLFWD) String() string {
	panic("PORKBUN_URLFWD should not be used directly.  It is a placeholder for the provider to substitute the correct type (URL or URL301).")
}

func MakePORKBUN_URLFWD(origin string, args ...any) (PORKBUN_URLFWD, error) {
	panic("PORKBUN_URLFWD should not be used directly.  It is a placeholder for the provider to substitute the correct type (URL or URL301).")
}
