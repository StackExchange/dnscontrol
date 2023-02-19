package recorddb

import "github.com/StackExchange/dnscontrol/v3/models"

// Functions that make it easier to deal with
// a group of records.
//

type RecordDB = struct {
	labelAndTypeMap map[models.RecordKey]struct{}
}

func NewFromRecords(recs models.Records) *RecordDB {
	result := &RecordDB{}

	result.labelAndTypeMap = make(map[models.RecordKey]struct{}, len(recs))
	for _, rec := range recs {
		result.labelAndTypeMap[rec.Key()] = struct{}{}
	}

	return result
}

// ContainsLT returns true if recdb contains rec. Matching is done
// on the record's label and type (i.e. the RecordKey)
//func (recdb RecordDB) ContainsLT(rec *models.RecordConfig) bool {
//	_, ok := recdb.labelAndTypeMap[rec.Key()]
//	return ok
//}
