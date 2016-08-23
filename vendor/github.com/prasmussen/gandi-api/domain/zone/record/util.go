package record

import (
    "github.com/prasmussen/gandi-api/util"
)


func ToRecordInfo(res map[string]interface{}) *RecordInfo {
    return &RecordInfo{
        Id: util.ToInt64(res["id"]),
        Name: util.ToString(res["name"]),
        Ttl: util.ToInt64(res["ttl"]),
        Type: util.ToString(res["type"]),
        Value: util.ToString(res["value"]),
    }
}
