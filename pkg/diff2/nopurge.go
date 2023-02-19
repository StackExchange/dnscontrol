package diff2

func processPurge(instructions ChangeList, nopurge bool) ChangeList {

	//	if nopurge {
	//		return instructions
	//	}
	//
	//	// TODO(tlim): This can probably be done without allocations but it
	//	// works and I won't want to prematurely optimize.
	//
	//	var msgs []string
	//
	//	newinstructions := make(ChangeList, 0, len(instructions))
	//	for _, j := range instructions {
	//		if j.Type == DELETE {
	//			msgs = append(msgs, j.Msgs...)
	//			continue
	//		}
	//		newinstructions = append(newinstructions, j)
	//	}
	//
	//	// Report what would have been purged
	//	if len(msgs) != 0 {
	//		for i := range msgs {
	//			msgs[i] = "NO_PURGE: Skipping " + msgs[i]
	//		}
	//		msgs = append([]string{"NO_PURGE Activated! Skipping these actions:"}, msgs...)
	//		newinstructions = append(newinstructions, Change{
	//			Type:       REPORT,
	//			Msgs:       msgs,
	//			MsgsJoined: strings.Join(msgs, "\n"),
	//		})
	//	}
	//
	//return newinstructions

	return instructions
}
