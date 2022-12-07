package diff2

func processPurge(instructions ChangeList, nopurge bool) ChangeList {

	if nopurge {
		return instructions
	}

	// TODO(tlim): This can probably be done without allocations but it
	// works and I won't want to prematurely optimize.

	newinstructions := make(ChangeList, 0, len(instructions))
	for _, j := range instructions {
		if j.Type == DELETE {
			continue
		}
		newinstructions = append(newinstructions, j)
	}

	return instructions

}
