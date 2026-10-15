package soapclient

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const serviceNamespace = "http://example.com/customer"
const soapAction = serviceNamespace + "/ICustomerService/GetCustomerById"

var ErrNotFound = errors.New("customer not found")

// Customer is the data contract exposed by the WCF service.
type Customer struct {
	CustomerID int    `json:"customerId" xml:"CustomerId"`
	Name       string `json:"name" xml:"Name"`
	Email      string `json:"email" xml:"Email"`
}

type Client struct {
	endpoint   string
	httpClient *http.Client
}

func New(endpoint string, timeout time.Duration) *Client {
	return &Client{
		endpoint:   endpoint,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) CustomerByID(ctx context.Context, id int) (Customer, error) {
	soapRequest := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
  <s:Body>
    <GetCustomerById xmlns="%s">
      <customerId>%s</customerId>
    </GetCustomerById>
  </s:Body>
</s:Envelope>`, serviceNamespace, strconv.Itoa(id))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewBufferString(soapRequest))
	if err != nil {
		return Customer{}, fmt.Errorf("create WCF request: %w", err)
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", `"`+soapAction+`"`)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Customer{}, fmt.Errorf("call WCF endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // cap response size at 2 MiB
	if err != nil {
		return Customer{}, fmt.Errorf("read WCF response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Customer{}, fmt.Errorf("WCF returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var envelope soapEnvelope
	if err := xml.Unmarshal(body, &envelope); err != nil {
		return Customer{}, fmt.Errorf("decode WCF SOAP XML: %w", err)
	}
	if envelope.Body.Fault != nil {
		return Customer{}, fmt.Errorf("WCF SOAP fault: %s", envelope.Body.Fault.String)
	}
	result := envelope.Body.GetCustomerByIDResponse.Result
	if result == nil || result.Nil == "true" || result.Nil == "1" {
		return Customer{}, ErrNotFound
	}

	return Customer{
		CustomerID: result.CustomerID,
		Name:       result.Name,
		Email:      result.Email,
	}, nil
}

type soapEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		GetCustomerByIDResponse struct {
			Result *customerResult `xml:"GetCustomerByIdResult"`
		} `xml:"GetCustomerByIdResponse"`
		Fault *soapFault `xml:"Fault"`
	} `xml:"Body"`
}

type customerResult struct {
	Nil        string `xml:"nil,attr"`
	CustomerID int    `xml:"CustomerId"`
	Name       string `xml:"Name"`
	Email      string `xml:"Email"`
}

type soapFault struct {
	Code   string `xml:"faultcode"`
	String string `xml:"faultstring"`
}
