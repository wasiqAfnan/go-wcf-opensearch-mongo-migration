using System.Runtime.Serialization;

namespace CustomerWcfService.Models;

[DataContract]
public sealed class Customer
{
    [DataMember(Order = 1)]
    public int CustomerId { get; set; }

    [DataMember(Order = 2)]
    public string Name { get; set; } = string.Empty;

    [DataMember(Order = 3)]
    public string? Email { get; set; }
}
