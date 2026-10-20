using System.Text.Json;
using CoreWCF;
using CustomerWcfService.Contracts;
using CustomerWcfService.Models;

namespace CustomerWcfService.Services;

[ServiceBehavior(IncludeExceptionDetailInFaults = true)]
public sealed class InvoiceSearchService(HttpClient httpClient, IConfiguration configuration)
    : IInvoiceSearchService
{
    private readonly string _openSearchUrl = configuration["OpenSearch:Url"] ?? "http://localhost:9200";

    public List<InvoiceDocumentContract> GetInvoicesByCustomerId(int customerId)
    {
        if (customerId <= 0)
        {
            throw new FaultException("CustomerId must be positive.");
        }

        var searchEndpoint = $"{_openSearchUrl.TrimEnd('/')}/customer_invoices/_search?q=customer_id:{customerId}&size=100";
        
        HttpResponseMessage response;
        try
        {
            response = httpClient.GetAsync(searchEndpoint).GetAwaiter().GetResult();
        }
        catch (Exception ex)
        {
            throw new FaultException($"Failed to query OpenSearch: {ex.Message}");
        }

        using (response)
        {
            if (!response.IsSuccessStatusCode)
            {
                return [];
            }

            using var jsonStream = response.Content.ReadAsStream();
            using var doc = JsonDocument.Parse(jsonStream);

            if (!doc.RootElement.TryGetProperty("hits", out var hitsContainer) ||
                !hitsContainer.TryGetProperty("hits", out var hits))
            {
                return [];
            }

            var results = new List<InvoiceDocumentContract>();
            foreach (var hit in hits.EnumerateArray())
            {
                if (!hit.TryGetProperty("_source", out var src))
                {
                    continue;
                }

                var invoice = new InvoiceDocumentContract
                {
                    InvoiceId = src.TryGetProperty("invoice_id", out var idElem) ? idElem.GetString() ?? "" : "",
                    CustomerId = src.TryGetProperty("customer_id", out var custElem) ? custElem.GetInt32() : 0,
                    IssueDate = src.TryGetProperty("issue_date", out var dateElem) ? dateElem.GetString() ?? "" : "",
                    Status = src.TryGetProperty("status", out var statusElem) ? statusElem.GetString() ?? "" : "",
                    Currency = src.TryGetProperty("currency", out var currElem) ? currElem.GetString() ?? "" : "",
                    TaxAmount = src.TryGetProperty("tax_amount", out var taxElem) ? taxElem.GetDecimal() : 0m,
                    TotalAmount = src.TryGetProperty("total_amount", out var totalElem) ? totalElem.GetDecimal() : 0m
                };

                if (src.TryGetProperty("items", out var itemsElem) && itemsElem.ValueKind == JsonValueKind.Array)
                {
                    foreach (var item in itemsElem.EnumerateArray())
                    {
                        invoice.Items.Add(new InvoiceLineItemContract
                        {
                            Sku = item.TryGetProperty("sku", out var skuElem) ? skuElem.GetString() ?? "" : "",
                            Description = item.TryGetProperty("description", out var descElem) ? descElem.GetString() ?? "" : "",
                            Quantity = item.TryGetProperty("quantity", out var qtyElem) ? qtyElem.GetInt32() : 0,
                            UnitPrice = item.TryGetProperty("unit_price", out var priceElem) ? priceElem.GetDecimal() : 0m
                        });
                    }
                }

                results.Add(invoice);
            }

            return results;
        }
    }
}

