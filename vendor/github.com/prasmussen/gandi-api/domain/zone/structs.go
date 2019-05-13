package zone

import (
	"time"
)

type ZoneInfoBase struct {
	DateUpdated time.Time
	Id          int64
	Name        string
	Public      bool
	Version     int64
}

type ZoneInfoExtra struct {
	Domains  int64
	Owner    string
	Versions []int64
}

type ZoneInfo struct {
	*ZoneInfoBase
	*ZoneInfoExtra
}
