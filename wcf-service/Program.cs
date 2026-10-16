using CoreWCF;
using CoreWCF.Configuration;
using CoreWCF.Description;
using CustomerWcfService.Contracts;
using CustomerWcfService.Data;
using CustomerWcfService.Services;

// This is the main entry point for the WCF service.
// It is responsible for configuring the service and hosting it.
var builder = WebApplication.CreateBuilder(args);
builder.WebHost.UseUrls(
    Environment.GetEnvironmentVariable("ASPNETCORE_URLS") ?? "http://localhost:8081");

// read the connection string from the environment variable for the database
var connectionString = builder.Configuration.GetConnectionString("AppDb");
if (string.IsNullOrWhiteSpace(connectionString))
{
    throw new InvalidOperationException(
        "Set the ConnectionStrings__AppDb environment variable before starting the WCF service.");
}

// configure the service
builder.Services.AddSingleton(new CustomerRepository(connectionString));
builder.Services.AddTransient<CustomerService>();
builder.Services.AddServiceModelServices();
builder.Services.AddServiceModelMetadata();
builder.Services.AddSingleton<IServiceBehavior, UseRequestHeadersForMetadataAddressBehavior>();

var app = builder.Build();

app.MapGet("/healthz", () => Results.Ok(new { status = "ok" }));

// Comfigure CoreWCF SOAP Endpoints 
app.UseServiceModel(serviceBuilder =>
{
    serviceBuilder.AddService<CustomerService>();
    serviceBuilder.AddServiceEndpoint<CustomerService, ICustomerService>(
        new BasicHttpBinding(), "/CustomerService");
});

// configure the metadata endpoint
var metadataBehavior = app.Services.GetRequiredService<ServiceMetadataBehavior>();
metadataBehavior.HttpGetEnabled = true;

// running the service
app.Run();
