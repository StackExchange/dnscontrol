package privatetypesrdata

type AZURE_ALIAS struct {
	Target string
}

func (rd AZURE_ALIAS) Len() int {
	return len(rd.Target) + 1
}

func (rd AZURE_ALIAS) String() string {
	return rd.Target
}
