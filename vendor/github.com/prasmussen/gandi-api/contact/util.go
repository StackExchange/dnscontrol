package contact

import (
    "github.com/prasmussen/gandi-api/util"
)

func toBalanceInformation(res map[string]interface{}) *BalanceInformation {
    return &BalanceInformation{
        AnnualBalance: util.ToString(res["annual_balance"]),
        Grid: util.ToString(res["grid"]),
        OutstandingAmount: util.ToFloat64(res["outstanding_amount"]),
        Prepaid: toPrepaidInformation(util.ToXmlrpcStruct(res["prepaid"])),
    }
}

func toPrepaidInformation(res map[string]interface{}) *PrepaidInformation {
    return &PrepaidInformation{
        Id: util.ToInt64(res["id"]),
        Amount: util.ToString(res["amount"]),
        Currency: util.ToString(res["currency"]),
        DateCreated: util.ToTime(res["date_created"]),
        DateUpdated: util.ToTime(res["date_updated"]),
    }
}

func toContactInformation(res map[string]interface{}) *ContactInformation {
    return &ContactInformation{
        Firstname: util.ToString(res["given"]),
        Lastname: util.ToString(res["family"]),
        Email: util.ToString(res["email"]),
        Address: util.ToString(res["streetaddr"]),
        Zipcode: util.ToString(res["zip"]),
        City: util.ToString(res["city"]),
        Country: util.ToString(res["country"]),
        Phone: util.ToString(res["phone"]),
        ContactType: util.ToInt64(res["type"]),
        Handle: util.ToString(res["handle"]),
    }
}
