package yd4b

import (
	"net/http"
)

func Do(c *Client, req *http.Request) (*http.Response, error) {
	return c.do(req)
}

func GetDo(c *Client) func(req *http.Request) (*http.Response, error) {
	return c.doFunc
}

func GetECUID(c *Client) string {
	return c.ecuid
}

func GetOrigin(c *Client) string {
	return c.origin
}

var NewSearchcodeRequest = newSearchcodeRequest

type AddressRequest = addressRequest
type AddressRequestOption = addressRequestOption

func NewAddressRequest(opts ...addressRequestOption) addressRequest {
	return newAddressRequest(opts...)
}
