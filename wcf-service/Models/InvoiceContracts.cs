using System.Runtime.Serialization;

namespace CustomerWcfService.Models;

[DataContract(Namespace = "http://example.com/invoices")]
public sealed class InvoiceLineItemContract
{
    [DataMember(Order = 1)]
    public string Sku { get; set; } = string.Empty;

    [DataMember(Order = 2)]
    public string Description { get; set; } = string.Empty;

    [DataMember(Order = 3)]
    public int Quantity { get; set; }

    [DataMember(Order = 4)]
    public decimal UnitPrice { get; set; }
}

[DataContract(Namespace = "http://example.com/invoices")]
public sealed class InvoiceDocumentContract
{
    [DataMember(Order = 1)]
    public string InvoiceId { get; set; } = string.Empty;

    [DataMember(Order = 2)]
    public int CustomerId { get; set; }

    [DataMember(Order = 3)]
    public string IssueDate { get; set; } = string.Empty;

    [DataMember(Order = 4)]
    public string Status { get; set; } = string.Empty;

    [DataMember(Order = 5)]
    public string Currency { get; set; } = string.Empty;

    [DataMember(Order = 6)]
    public List<InvoiceLineItemContract> Items { get; set; } = [];

    [DataMember(Order = 7)]
    public decimal TaxAmount { get; set; }

    [DataMember(Order = 8)]
    public decimal TotalAmount { get; set; }
}

