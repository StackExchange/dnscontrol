package client

import "github.com/kolo/xmlrpc"

const (
	Production SystemType = iota
	Testing
)

type SystemType int

func (self SystemType) Url() string {
	if self == Production {
		return "https://rpc.gandi.net/xmlrpc/"
	}
	return "https://rpc.ote.gandi.net/xmlrpc/"
}

type Client struct {
	Key string
	Url string
}

func New(apiKey string, system SystemType) *Client {
	return &Client{
		Key: apiKey,
		Url: system.Url(),
	}
}

func (self *Client) Call(serviceMethod string, args []interface{}, reply interface{}) error {
	rpc, err := xmlrpc.NewClient(self.Url, nil)
	if err != nil {
		return err
	}
	return rpc.Call(serviceMethod, args, reply)
}
