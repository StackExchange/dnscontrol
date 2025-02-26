package main

import (
	"strconv"

	"github.com/StackExchange/dnscontrol/v4/models"
)

func a(name string, a string) *models.RecordConfig {

	rdata, err := models.ParseA([]string{a}, "", "**current-domain**")
	if err != nil {
		panic(err)
	}
	return models.MustCreateRecord(name, rdata, nil, 300, "**current-domain**")
}

func ns(name string, ns string) *models.RecordConfig {

	rdata, err := models.ParseNS([]string{ns}, "", "**current-domain**")
	if err != nil {
		panic(err)
	}
	return models.MustCreateRecord(name, rdata, nil, 300, "**current-domain**")
}

func cname(name string, target string) *models.RecordConfig {

	rdata, err := models.ParseCNAME([]string{target}, "", "**current-domain**")
	if err != nil {
		panic(err)
	}
	return models.MustCreateRecord(name, rdata, nil, 300, "**current-domain**")
}

func ptr(name string, ptr string) *models.RecordConfig {

	rdata, err := models.ParsePTR([]string{ptr}, "", "**current-domain**")
	if err != nil {
		panic(err)
	}
	return models.MustCreateRecord(name, rdata, nil, 300, "**current-domain**")
}

func mx(name string, preference uint16, mx string) *models.RecordConfig {
	spreference := strconv.Itoa(int(preference))

	rdata, err := models.ParseMX([]string{spreference, mx}, "", "**current-domain**")
	if err != nil {
		panic(err)
	}
	return models.MustCreateRecord(name, rdata, nil, 300, "**current-domain**")
}

func aaaa(name string, aaaa string) *models.RecordConfig {

	rdata, err := models.ParseAAAA([]string{aaaa}, "", "**current-domain**")
	if err != nil {
		panic(err)
	}
	return models.MustCreateRecord(name, rdata, nil, 300, "**current-domain**")
}

func srv(name string, priority uint16, weight uint16, port uint16, target string) *models.RecordConfig {
	spriority := strconv.Itoa(int(priority))
	sweight := strconv.Itoa(int(weight))
	sport := strconv.Itoa(int(port))

	rdata, err := models.ParseSRV([]string{spriority, sweight, sport, target}, "", "**current-domain**")
	if err != nil {
		panic(err)
	}
	return models.MustCreateRecord(name, rdata, nil, 300, "**current-domain**")
}

func naptr(name string, order uint16, preference uint16, flags string, service string, regexp string, replacement string) *models.RecordConfig {
	sorder := strconv.Itoa(int(order))
	spreference := strconv.Itoa(int(preference))

	rdata, err := models.ParseNAPTR([]string{sorder, spreference, flags, service, regexp, replacement}, "", "**current-domain**")
	if err != nil {
		panic(err)
	}
	return models.MustCreateRecord(name, rdata, nil, 300, "**current-domain**")
}

func ds(name string, keytag uint16, algorithm uint8, digesttype uint8, digest string) *models.RecordConfig {
	skeytag := strconv.Itoa(int(keytag))
	salgorithm := strconv.Itoa(int(algorithm))
	sdigesttype := strconv.Itoa(int(digesttype))

	rdata, err := models.ParseDS([]string{skeytag, salgorithm, sdigesttype, digest}, "", "**current-domain**")
	if err != nil {
		panic(err)
	}
	return models.MustCreateRecord(name, rdata, nil, 300, "**current-domain**")
}

func dnskey(name string, flags uint16, protocol uint8, algorithm uint8, publickey string) *models.RecordConfig {
	sflags := strconv.Itoa(int(flags))
	sprotocol := strconv.Itoa(int(protocol))
	salgorithm := strconv.Itoa(int(algorithm))

	rdata, err := models.ParseDNSKEY([]string{sflags, sprotocol, salgorithm, publickey}, "", "**current-domain**")
	if err != nil {
		panic(err)
	}
	return models.MustCreateRecord(name, rdata, nil, 300, "**current-domain**")
}

func caa(name string, flag uint8, tag string, value string) *models.RecordConfig {
	sflag := strconv.Itoa(int(flag))

	rdata, err := models.ParseCAA([]string{sflag, tag, value}, "", "**current-domain**")
	if err != nil {
		panic(err)
	}
	return models.MustCreateRecord(name, rdata, nil, 300, "**current-domain**")
}

func cfsingleredirect(srname string, code uint16, srwhen string, srthen string) *models.RecordConfig {
	name := srname
	scode := strconv.Itoa(int(code))

	rdata, err := models.ParseCFSINGLEREDIRECT([]string{srname, scode, srwhen, srthen}, "", "**current-domain**")
	if err != nil {
		panic(err)
	}
	return models.MustCreateRecord(name, rdata, nil, 300, "**current-domain**")
}
