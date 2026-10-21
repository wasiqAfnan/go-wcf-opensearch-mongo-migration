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

const customerNamespace = "http://example.com/customer"
const customerSoapAction = customerNamespace + "/ICustomerService/GetCustomerById"

const invoiceNamespace = "http://example.com/invoices"
const invoiceSoapAction = invoiceNamespace + "/IInvoiceSearchService/GetInvoicesByCustomerId"

var ErrNotFound = errors.New("customer not found")

// Customer is the data contract exposed by the WCF CustomerService.
type Customer struct {
	CustomerID int    `json:"customerId" xml:"CustomerId"`
	Name       string `json:"name" xml:"Name"`
	Email      string `json:"email" xml:"Email"`
}

// InvoiceLineItem is the line item data contract in the invoice response.
type InvoiceLineItem struct {
	Sku         string  `json:"sku" xml:"Sku"`
	Description string  `json:"description" xml:"Description"`
	Quantity    int     `json:"quantity" xml:"Quantity"`
	UnitPrice   float64 `json:"unitPrice" xml:"UnitPrice"`
}

// InvoiceDocument is the invoice data contract exposed by the WCF InvoiceSearchService.
type InvoiceDocument struct {
	InvoiceId   string            `json:"invoiceId" xml:"InvoiceId"`
	CustomerId  int               `json:"customerId" xml:"CustomerId"`
	IssueDate   string            `json:"issueDate" xml:"IssueDate"`
	Status      string            `json:"status" xml:"Status"`
	Currency    string            `json:"currency" xml:"Currency"`
	Items       []InvoiceLineItem `json:"items" xml:"Items>InvoiceLineItemContract"`
	TaxAmount   float64           `json:"taxAmount" xml:"TaxAmount"`
	TotalAmount float64           `json:"totalAmount" xml:"TotalAmount"`
}

type Client struct {
	customerURL string
	invoiceURL  string
	httpClient  *http.Client
}

// New creates a new SOAP client. endpoint can be the base WCF URL (e.g. http://localhost:8081)
// or the legacy CustomerService endpoint URL.
func New(endpoint string, timeout time.Duration) *Client {
	base := strings.TrimRight(endpoint, "/")
	base = strings.TrimSuffix(base, "/CustomerService")
	base = strings.TrimSuffix(base, "/InvoiceSearchService")

	return &Client{
		customerURL: base + "/CustomerService",
		invoiceURL:  base + "/InvoiceSearchService",
		httpClient:  &http.Client{Timeout: timeout},
	}
}

// CustomerByID retrieves customer master information from CoreWCF CustomerService.
func (c *Client) CustomerByID(ctx context.Context, id int) (Customer, error) {
	soapRequest := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
  <s:Body>
    <GetCustomerById xmlns="%s">
      <customerId>%s</customerId>
    </GetCustomerById>
  </s:Body>
</s:Envelope>`, customerNamespace, strconv.Itoa(id))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.customerURL, bytes.NewBufferString(soapRequest))
	if err != nil {
		return Customer{}, fmt.Errorf("create WCF customer request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", `"`+customerSoapAction+`"`)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Customer{}, fmt.Errorf("call WCF customer endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return Customer{}, fmt.Errorf("read WCF customer response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Customer{}, fmt.Errorf("WCF returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var envelope customerSoapEnvelope
	if err := xml.Unmarshal(body, &envelope); err != nil {
		return Customer{}, fmt.Errorf("decode WCF customer SOAP XML: %w", err)
	}
	if envelope.Body.Fault != nil {
		return Customer{}, fmt.Errorf("WCF customer SOAP fault: %s", envelope.Body.Fault.String)
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

// InvoicesByCustomerID retrieves invoices for a customer from CoreWCF InvoiceSearchService.
func (c *Client) InvoicesByCustomerID(ctx context.Context, customerID int) ([]InvoiceDocument, error) {
	soapRequest := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
  <s:Body>
    <GetInvoicesByCustomerId xmlns="%s">
      <customerId>%d</customerId>
    </GetInvoicesByCustomerId>
  </s:Body>
</s:Envelope>`, invoiceNamespace, customerID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.invoiceURL, bytes.NewBufferString(soapRequest))
	if err != nil {
		return nil, fmt.Errorf("create WCF invoices request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", `"`+invoiceSoapAction+`"`)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call WCF invoices endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
	if err != nil {
		return nil, fmt.Errorf("read WCF invoices response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("WCF returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var envelope invoiceSoapEnvelope
	if err := xml.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("decode WCF invoices SOAP XML: %w", err)
	}
	if envelope.Body.Fault != nil {
		return nil, fmt.Errorf("WCF invoices SOAP fault: %s", envelope.Body.Fault.String)
	}

	invoices := envelope.Body.GetInvoicesByCustomerIdResponse.Result.Invoices
	if invoices == nil {
		invoices = []InvoiceDocument{}
	}
	return invoices, nil
}

type customerSoapEnvelope struct {
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

type invoiceSoapEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		GetInvoicesByCustomerIdResponse struct {
			Result struct {
				Invoices []InvoiceDocument `xml:"InvoiceDocumentContract"`
			} `xml:"GetInvoicesByCustomerIdResult"`
		} `xml:"GetInvoicesByCustomerIdResponse"`
		Fault *soapFault `xml:"Fault"`
	} `xml:"Body"`
}

type soapFault struct {
	Code   string `xml:"faultcode"`
	String string `xml:"faultstring"`
}
