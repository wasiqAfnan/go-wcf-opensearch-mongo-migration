using CoreWCF;
using CustomerWcfService.Models;

namespace CustomerWcfService.Contracts;

[ServiceContract(Namespace = "http://example.com/customer")]
public interface ICustomerService
{
    [OperationContract]
    Customer? GetCustomerById(int customerId);
}
