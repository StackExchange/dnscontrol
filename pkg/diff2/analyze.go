package diff2

import (
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
)

func analyze(
	existing models.Records,
	desired models.Records,
) (
	creations, deletions, modifications []Change,
	err error,
) {

	dc := models.DomainConfig{
		Records: desired,
	}

	differ := diff.New(&dc)
	_, changeCS, delCS, modCS, err := differ.IncrementalDiff(existing)
	if err != nil {
		return nil, nil, nil, err
	}

	return changeSetToChange(changeCS), changeSetToChange(delCS), changeSetToChange(modCS), nil
}

func changeSetToChange(csl diff.Changeset) []Change {

	cl := make([]Change, len(csl))
	for i, c := range csl {
		cl[i] = Change{
			Old: c.Existing,
			New: c.Desired,
			Msg: c.String(),
		}
	}

	return cl
}
