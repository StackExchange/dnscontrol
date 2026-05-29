package privatetypesrdata

type AZURE_ALIAS struct {
	AliasType string
	Target    string
}

func (rd AZURE_ALIAS) Len() int {
	return len(rd.Target) + 1 + len(rd.AliasType)
}

func (rd AZURE_ALIAS) String() string {
	return rd.AliasType + " " + rd.Target
}
