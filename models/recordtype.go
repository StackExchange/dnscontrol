package models

// ChangeType changes the type of a record.
func (rc *RecordConfig) ChangeType(newType string, origin string) {

	rc.Type = newType

	if IsTypeUpgraded(newType) {
		rc.ImportFromLegacy(origin)
	}

}
