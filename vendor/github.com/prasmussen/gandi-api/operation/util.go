package operation

import (
	"github.com/prasmussen/gandi-api/util"
)

func ToOperationInfo(res map[string]interface{}) *OperationInfo {
	return &OperationInfo{
		DateCreated:      util.ToTime(res["date_created"]),
		DateStart:        util.ToTime(res["date_start"]),
		DateUpdated:      util.ToTime(res["date_updated"]),
		Eta:              util.ToString(res["eta"]),
		Id:               util.ToInt64(res["id"]),
		LastError:        util.ToString(res["last_error"]),
		SessionId:        util.ToInt64(res["session_id"]),
		Source:           util.ToString(res["source"]),
		Step:             util.ToString(res["step"]),
		Type:             util.ToString(res["type"]),
		OperationDetails: ToOperationDetails(util.ToXmlrpcStruct(res["infos"])),
		Params:           util.ToXmlrpcStruct(res["params"]),
	}
}

func ToOperationDetails(res map[string]interface{}) *OperationDetails {
	return &OperationDetails{
		Id:            util.ToString(res["id"]),
		Label:         util.ToString(res["label"]),
		ProductAction: util.ToString(res["product_action"]),
		ProductName:   util.ToString(res["product_name"]),
		ProductType:   util.ToString(res["product_type"]),
		Quantity:      util.ToInt64(res["quantity"]),
	}
}
