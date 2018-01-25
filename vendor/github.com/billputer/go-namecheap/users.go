package namecheap

import (
	"net/url"
)

const (
	usersGetPricing = "namecheap.users.getPricing"
)

type UsersGetPricingResult struct {
	ProductType     string `xml:"Name,attr"`
	ProductCategory []struct {
		Name    string `xml:"Name,attr"`
		Product []struct {
			Name  string `xml:"Name,attr"`
			Price []struct {
				Duration     int     `xml:"Duration,attr"`
				DurationType string  `xml:"DurationType,attr"`
				Price        float32 `xml:"Price,attr"`
				RegularPrice float32 `xml:"RegularPrice,attr"`
				YourPrice    float32 `xml:"YourPrice,attr"`
				CouponPrice  float32 `xml:"CouponPrice,attr"`
				Currency     string  `xml:"Currency,attr"`
			} `xml:"Price"`
		} `xml:"Product"`
	} `xml:"ProductCategory"`
}

func (client *Client) UsersGetPricing(productType string) ([]UsersGetPricingResult, error) {
	requestInfo := &ApiRequest{
		command: usersGetPricing,
		method:  "GET",
		params:  url.Values{},
	}

	requestInfo.params.Set("ProductType", productType)
	resp, err := client.do(requestInfo)
	if err != nil {
		return nil, err
	}

	return resp.UsersGetPricing, nil
}
