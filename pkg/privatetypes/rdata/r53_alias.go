package privatetypesrdata

type R53_ALIAS struct {
	AliasType, Target, EvalTargetHealth, ZoneID string
}

func (rr R53_ALIAS) Len() int {
	return len(rr.AliasType) +
		1 + len(rr.Target) +
		1 + len(rr.EvalTargetHealth) +
		1 + len(rr.ZoneID)
}

func (rd R53_ALIAS) String() string {
	return rd.AliasType +
		" " + rd.Target +
		" " + rd.EvalTargetHealth +
		" " + rd.ZoneID
}
