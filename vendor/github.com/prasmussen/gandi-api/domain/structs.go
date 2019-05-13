package domain

import (
	"time"
)

type DomainInfoBase struct {
	AuthInfo             string
	DateCreated          time.Time
	DateRegistryCreation time.Time
	DateRegistryEnd      time.Time
	DateUpdated          time.Time
	Fqdn                 string
	Id                   int64
	Status               []string
	Tld                  string
}

type DomainInfoExtra struct {
	DateDelete           time.Time
	DateHoldBegin        time.Time
	DateHoldEnd          time.Time
	DatePendingDeleteEnd time.Time
	DateRenewBegin       time.Time
	DateRestoreEnd       time.Time
	Nameservers          []string
	Services             []string
	ZoneId               int64
	Autorenew            *AutorenewInfo
	Contacts             *ContactInfo
}

type DomainInfo struct {
	*DomainInfoBase
	*DomainInfoExtra
}

type AutorenewInfo struct {
	Active        bool
	Contact       string
	Id            int64
	ProductId     int64
	ProductTypeId int64
}

type ContactInfo struct {
	Admin    *ContactDetails
	Bill     *ContactDetails
	Owner    *ContactDetails
	Reseller *ContactDetails
	Tech     *ContactDetails
}

type ContactDetails struct {
	Handle string
	Id     int64
}
