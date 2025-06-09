package none

//// toRc converts a native record (what is received from the API) to a RecordConfig.
//func toRc(domain string, r theSdkModule.NativeRecordType) (*models.RecordConfig, error) {
//}

//// toRc converts a RecordConfig to a native record (what is received from the API).
//func toNative(rc *models.RecordConfig) theSdkModule.NativeRecordType {
//}

/* Or...

If your provider stores all records at the same label (or label+type) in one type (often called a RecordSet).

func toRcs(n theSdkModule.NativeRecordSet, origin string) (rcs []*models.RecordConfig, err error) {}

func toNative(rcs []*models.RecordConfig, origin string) []theSdkModule.NativeRecordSet {}

*/
