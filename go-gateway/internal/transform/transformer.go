package transform

import (
	"time"

	"example.com/go-wcf-learning/go-gateway/internal/soapclient"
)

type TransformedInvoice struct {
	ID             string         `bson:"_id" json:"id"`
	CustomerID     int            `bson:"customerId" json:"customerId"`
	Customer       CustomerInfo   `bson:"customer" json:"customer"`
	Invoice        InvoiceDetails `bson:"invoice" json:"invoice"`
	MigrationAudit AuditInfo      `bson:"migrationAudit" json:"migrationAudit"`
}

type CustomerInfo struct {
	Name  string `bson:"name" json:"name"`
	Email string `bson:"email" json:"email"`
}

type InvoiceDetails struct {
	IssueDate   string     `bson:"issueDate" json:"issueDate"`
	Status      string     `bson:"status" json:"status"`
	Currency    string     `bson:"currency" json:"currency"`
	TotalAmount float64    `bson:"totalAmount" json:"totalAmount"`
	TaxAmount   float64    `bson:"taxAmount" json:"taxAmount"`
	ItemCount   int        `bson:"itemCount" json:"itemCount"`
	LineItems   []LineItem `bson:"lineItems" json:"lineItems"`
}

type LineItem struct {
	SKU         string  `bson:"sku" json:"sku"`
	Description string  `bson:"description" json:"description"`
	Quantity    int     `bson:"quantity" json:"quantity"`
	UnitPrice   float64 `bson:"unitPrice" json:"unitPrice"`
}

type AuditInfo struct {
	MigratedAt      time.Time `bson:"migratedAt" json:"migratedAt"`
	SourceSystem    string    `bson:"sourceSystem" json:"sourceSystem"`
	PipelineVersion string    `bson:"pipelineVersion" json:"pipelineVersion"`
	Verified        bool      `bson:"verified" json:"verified"`
}

func Transform(customer soapclient.Customer, invoice soapclient.InvoiceDocument) TransformedInvoice {
	items := make([]LineItem, len(invoice.Items))
	for i, item := range invoice.Items {
		items[i] = LineItem{
			SKU:         item.Sku,
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		}
	}

	return TransformedInvoice{
		ID:         invoice.InvoiceId,
		CustomerID: customer.CustomerID,
		Customer: CustomerInfo{
			Name:  customer.Name,
			Email: customer.Email,
		},
		Invoice: InvoiceDetails{
			IssueDate:   invoice.IssueDate,
			Status:      invoice.Status,
			Currency:    invoice.Currency,
			TotalAmount: invoice.TotalAmount,
			TaxAmount:   invoice.TaxAmount,
			ItemCount:   len(items),
			LineItems:   items,
		},
		MigrationAudit: AuditInfo{
			MigratedAt:      time.Now().UTC(),
			SourceSystem:    "MSSQL_AND_OPENSEARCH_VIA_COREWCF",
			PipelineVersion: "1.0.0",
			Verified:        true,
		},
	}
}

