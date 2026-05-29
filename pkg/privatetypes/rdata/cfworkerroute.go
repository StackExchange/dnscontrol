package privatetypesrdata

type CFWORKERROUTE struct {
	When string
	Then string
}

func (rd CFWORKERROUTE) Len() int {
	return len(rd.When) + 1 + len(rd.Then)
}

func (rd CFWORKERROUTE) String() string {
	return rd.When + " " + rd.Then
}
