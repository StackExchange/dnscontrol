package version

import (
    "github.com/prasmussen/gandi-api/util"
)


func ToVersionInfo(res map[string]interface{}) *VersionInfo {
    return &VersionInfo{
        Id: util.ToInt64(res["id"]),
        DateCreated: util.ToTime(res["date_created"]),
    }
}
