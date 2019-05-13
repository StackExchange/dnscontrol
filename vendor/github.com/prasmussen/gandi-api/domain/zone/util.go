package zone

import (
	"github.com/prasmussen/gandi-api/util"
)

func ToZoneInfoBase(res map[string]interface{}) *ZoneInfoBase {
	return &ZoneInfoBase{
		DateUpdated: util.ToTime(res["date_updated"]),
		Id:          util.ToInt64(res["id"]),
		Name:        util.ToString(res["name"]),
		Public:      util.ToBool(res["public"]),
		Version:     util.ToInt64(res["version"]),
	}
}

func ToZoneInfoExtra(res map[string]interface{}) *ZoneInfoExtra {
	return &ZoneInfoExtra{
		Domains:  util.ToInt64(res["domains"]),
		Owner:    util.ToString(res["owner"]),
		Versions: util.ToIntSlice(util.ToInterfaceSlice(res["versions"])),
	}
}

func ToZoneInfo(res map[string]interface{}) *ZoneInfo {
	return &ZoneInfo{
		ToZoneInfoBase(res),
		ToZoneInfoExtra(res),
	}
}
