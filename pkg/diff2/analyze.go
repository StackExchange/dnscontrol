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

	differ := diff.New(dc)
	_, creates, dels, modifications, err := differ.IncrementalDiff(existing)
	if err != nil {
		return nil, err
	}

	return nil, nil, nil, nil
}
