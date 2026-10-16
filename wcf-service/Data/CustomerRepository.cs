using System.Data;
using CustomerWcfService.Models;
using Microsoft.Data.SqlClient;

namespace CustomerWcfService.Data;

// This class owns SQL Server access. It knows nothing about SOAP or HTTP.
public sealed class CustomerRepository(string connectionString)
{
    public Customer? GetById(int customerId)
    {   
        // create a command and connection to execute the stored procedure
        using var connection = new SqlConnection(connectionString);
        using var command = new SqlCommand("dbo.GetCustomerById", connection)
        {
            CommandType = CommandType.StoredProcedure,
            CommandTimeout = 5
        };
        command.Parameters.Add("@CustomerId", SqlDbType.Int).Value = customerId;

        // execute the command
        connection.Open();

        // read the result
        using var reader = command.ExecuteReader(CommandBehavior.SingleRow);
        if (!reader.Read())
        {
            return null;
        }

        // map the result to a Customer object
        var emailOrdinal = reader.GetOrdinal("Email");
        return new Customer
        {
            CustomerId = reader.GetInt32(reader.GetOrdinal("CustomerId")),
            Name = reader.GetString(reader.GetOrdinal("Name")),
            Email = reader.IsDBNull(emailOrdinal) ? null : reader.GetString(emailOrdinal)
        };
    }
}
