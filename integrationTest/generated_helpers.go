package main

import (
	"strconv"

	"github.com/StackExchange/dnscontrol/v4/models"
)

func a(name string, a string) *models.RecordConfig {

	rdata, err := models.ParseA([]string{a}, "**current-domain**")
	if err != nil {
		panic(err)
	}
	return models.MustCreateRecord(name, rdata, nil, 300, "**current-domain**")
}

func mx(name string, preference uint16, mx string) *models.RecordConfig {
	spreference := strconv.Itoa(int(preference))

	rdata, err := models.ParseMX([]string{spreference, mx}, "**current-domain**")
	if err != nil {
		panic(err)
	}
	return models.MustCreateRecord(name, rdata, nil, 300, "**current-domain**")
}

func srv(name string, priority uint16, weight uint16, port uint16, target string) *models.RecordConfig {
	spriority := strconv.Itoa(int(priority))
	sweight := strconv.Itoa(int(weight))
	sport := strconv.Itoa(int(port))

	rdata, err := models.ParseSRV([]string{spriority, sweight, sport, target}, "**current-domain**")
	if err != nil {
		panic(err)
	}
	return models.MustCreateRecord(name, rdata, nil, 300, "**current-domain**")
}

func cname(name string, target string) *models.RecordConfig {

	rdata, err := models.ParseCNAME([]string{target}, "**current-domain**")
	if err != nil {
		panic(err)
	}
	return models.MustCreateRecord(name, rdata, nil, 300, "**current-domain**")
}

func cfsingleredirect(name string, srname string, code uint16, srwhen string, srthen string, srrrulesetid string, srrrulesetruleid string, srdisplay string, prwhen string, prthen string, prpriority int, prdisplay string) *models.RecordConfig {
	scode := strconv.Itoa(int(code))
	sprpriority := strconv.Itoa(prpriority)

	rdata, err := models.ParseCFSINGLEREDIRECT([]string{srname, scode, srwhen, srthen, srrrulesetid, srrrulesetruleid, srdisplay, prwhen, prthen, sprpriority, prdisplay}, "**current-domain**")
	if err != nil {
		panic(err)
	}
	return models.MustCreateRecord(name, rdata, nil, 300, "**current-domain**")
}
