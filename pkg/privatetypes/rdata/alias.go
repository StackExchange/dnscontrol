package privatetypesrdata

type ALIAS struct {
	Target string
}

func (rd ALIAS) Len() int {
	return len(rd.Target) + 1
}

func (rd ALIAS) String() string {
	return rd.Target
}
