using CoreWCF;
using CustomerWcfService.Models;

namespace CustomerWcfService.Contracts;

[ServiceContract(Namespace = "http://example.com/invoices")]
public interface IInvoiceSearchService
{
    [OperationContract]
    List<InvoiceDocumentContract> GetInvoicesByCustomerId(int customerId);
}

