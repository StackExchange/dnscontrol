package domain

import (
    "github.com/prasmussen/gandi-api/util"
)

func ToDomainInfoBase(res map[string]interface{}) *DomainInfoBase {
    return &DomainInfoBase{
        AuthInfo: util.ToString(res["authinfo"]),
        DateCreated: util.ToTime(res["date_created"]),
        DateRegistryCreation: util.ToTime(res["date_registry_creation"]),
        DateRegistryEnd: util.ToTime(res["date_registry_end"]),
        DateUpdated: util.ToTime(res["date_updated"]),
        Fqdn: util.ToString(res["fqdn"]),
        Id: util.ToInt64(res["id"]),
        Status: util.ToStringSlice(util.ToInterfaceSlice(res["status"])),
        Tld: util.ToString(res["tld"]),
    }
}

func ToDomainInfoExtra(res map[string]interface{}) *DomainInfoExtra {
    return &DomainInfoExtra{
        DateDelete: util.ToTime(res["date_delete"]),
        DateHoldBegin: util.ToTime(res["date_hold_begin"]),
        DateHoldEnd: util.ToTime(res["date_hold_end"]),
        DatePendingDeleteEnd: util.ToTime(res["date_pending_delete_end"]),
        DateRenewBegin: util.ToTime(res["date_renew_begin"]),
        DateRestoreEnd: util.ToTime(res["date_restore_end"]),
        Nameservers: util.ToStringSlice(util.ToInterfaceSlice(res["nameservers"])),
        Services: util.ToStringSlice(util.ToInterfaceSlice(res["services"])),
        ZoneId: util.ToInt64(res["zone_id"]),
        Autorenew: toAutorenewInfo(util.ToXmlrpcStruct(res["autorenew"])),
        Contacts: toContactInfo(util.ToXmlrpcStruct(res["contacts"])),
    }
}

func ToDomainInfo(res map[string]interface{}) *DomainInfo {
    return &DomainInfo{
        ToDomainInfoBase(res),
        ToDomainInfoExtra(res),
    }
}

func toAutorenewInfo(res map[string]interface{}) *AutorenewInfo {
    return &AutorenewInfo{
        Active: util.ToBool(res["active"]),
        Contact: util.ToString(res["contact"]),
        Id: util.ToInt64(res["id"]),
        ProductId: util.ToInt64(res["product_id"]),
        ProductTypeId: util.ToInt64(res["product_type_id"]),
    }
}

func toContactInfo(res map[string]interface{}) *ContactInfo {
    return &ContactInfo{
        Admin: toContactDetails(util.ToXmlrpcStruct(res["admin"])),
        Bill: toContactDetails(util.ToXmlrpcStruct(res["bill"])),
        Owner: toContactDetails(util.ToXmlrpcStruct(res["owner"])),
        Reseller: toContactDetails(util.ToXmlrpcStruct(res["reseller"])),
        Tech: toContactDetails(util.ToXmlrpcStruct(res["tech"])),
    }
}

func toContactDetails(res map[string]interface{}) *ContactDetails {
    return &ContactDetails{
        Handle: util.ToString(res["handle"]),
        Id: util.ToInt64(res["id"]),
    }
}
