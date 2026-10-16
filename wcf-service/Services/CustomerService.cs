using CoreWCF;
using CustomerWcfService.Contracts;
using CustomerWcfService.Data;
using CustomerWcfService.Models;

namespace CustomerWcfService.Services;

[ServiceBehavior(IncludeExceptionDetailInFaults = true)]
public sealed class CustomerService(CustomerRepository customers)
    : ICustomerService
{
    public Customer? GetCustomerById(int customerId)
    {
        if (customerId <= 0)
        {
            throw new FaultException(
                "Customer ID must be a positive integer.");
        }

        return customers.GetById(customerId);
    }
}