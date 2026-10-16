# Go → CoreWCF → SQL Server Learning Lab (Ubuntu 24.04)

This is the Ubuntu-runnable variant of the Go/WCF/MSSQL learning lab. It demonstrates one vertical slice:

`curl/HTTP -> Go JSON API -> SOAP/XML -> CoreWCF service -> ADO.NET -> SQL Server -> SOAP/XML -> Go JSON`

## Important platform distinction

The original version of this lab used traditional WCF on .NET Framework 4.8. That server cannot be built and hosted natively on Ubuntu. This Ubuntu edition uses **CoreWCF**, a community-led port that implements WCF-style service contracts and hosts SOAP endpoints on modern .NET/ASP.NET Core. The service contracts and SOAP concepts are intentionally similar, but CoreWCF is not identical to the old .NET Framework WCF runtime.

If your workplace's WCF service is a legacy .NET Framework service, you do not have to host it locally: run this Go gateway on Ubuntu and point `WCF_SOAP_URL` at the real, reachable service endpoint. To run the legacy service itself, use Windows or a Windows-hosted development environment.

## Stack

- Go 1.22+ (your project uses Go 1.22 syntax; Go 1.23+ is fine)
- .NET SDK 10
- CoreWCF HTTP/Primitives 1.9.1
- Microsoft.Data.SqlClient 7.1.1
- Docker Engine + Docker Compose plugin
- SQL Server 2022 Developer Linux container (for local learning; x86-64/amd64 host recommended)

## 1. Install/check prerequisites

Check Go:

```bash
go version
```

Install .NET 10 SDK on Ubuntu 24.04 from Ubuntu's package feed:

```bash
sudo apt update
sudo apt install -y dotnet-sdk-10.0
```

Check:

```bash
dotnet --info
```

Install Docker Engine using the [official Ubuntu installation instructions](https://docs.docker.com/engine/install/ubuntu/). Ensure the Docker Compose plugin is installed and Docker is running:

```bash
sudo systemctl enable --now docker
docker --version
docker compose version
sudo docker run --rm hello-world
```

If Docker commands report a permission error, either prefix each `docker` command in this guide with `sudo`, or add your account to the Docker group and then log out/in:

```bash
sudo usermod -aG docker "$USER"
```

Membership in the `docker` group effectively grants root-level privileges, so only do this on a machine where that is acceptable.

## 2. Start SQL Server in Docker

From the project root:

```bash
cp .env.example .env
nano .env
```

Replace the sample password with your own strong local development password. Do not use production credentials. Then start SQL Server:

```bash
docker compose up -d

docker ps

docker logs -f learningstack-sql
```

Wait for the SQL Server log to report it is ready for client connections, then press `Ctrl+C` to exit log-following (the container keeps running).

The Compose file binds port 1433 to `127.0.0.1`, so the database port is not exposed on all network interfaces. A named Docker volume persists SQL Server data if the container is recreated.

## 3. Create the database, table, sample row, and stored procedure

Copy and run the SQL script inside the SQL Server container:

```bash
docker cp database/setup.sql learningstack-sql:/tmp/setup.sql

docker exec -it learningstack-sql bash -lc '/opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -C -b -i /tmp/setup.sql'
```

This creates database `LearningStack`, `dbo.Customers`, sample customer `101`, and stored procedure `dbo.GetCustomerById`.

If the `sqlcmd` path is missing, inspect the installed tools with:

```bash
docker exec -it learningstack-sql bash -lc 'find /opt -type f -name sqlcmd 2>/dev/null'
```

Use the path returned by `find` in the setup command.

## 4. Run the CoreWCF service

Open **Terminal 1**. From the project root, load the local password and create the connection-string environment variable expected by .NET configuration:

```bash
set -a
source .env
set +a
export ConnectionStrings__AppDb="Server=127.0.0.1,1433;Database=LearningStack;User ID=sa;Password=${MSSQL_SA_PASSWORD};Encrypt=True;TrustServerCertificate=True"
dotnet run --project wcf-service/CustomerWcfService.csproj
```

`TrustServerCertificate=True` is included for this local development container only. Use proper TLS certificate validation and a least-privilege database login in real environments. The example uses `sa` for learning convenience, not as a production recommendation.

The service listens on `http://localhost:8081` by default. Test its health endpoint from another terminal:

```bash
curl -i http://localhost:8081/healthz
```

Check its WSDL/metadata:

```bash
curl -fsS 'http://localhost:8081/CustomerService?wsdl' | head
```

CoreWCF uses ASP.NET Core hosting. `ServiceContract`, `OperationContract`, the data contract, `BasicHttpBinding`, and SOAP remain the important ideas to focus on.

## 5. Run the Go API gateway

Open **Terminal 2**, change to the project root, and run:

```bash
cd go-gateway
WCF_SOAP_URL='http://localhost:8081/CustomerService' go run ./cmd/api
```

The Go API listens on port 8080.

Run its unit tests in another terminal:

```bash
cd go-gateway
go test ./...
```

## 6. Test the complete path

Open **Terminal 3**:

```bash
curl -i http://localhost:8080/healthz
curl -i http://localhost:8080/api/v1/customers/101
curl -i http://localhost:8080/api/v1/customers/9999
```

Expected successful JSON for customer 101:

```json
{
  "customerId": 101,
  "name": "Rahul Sharma",
  "email": "rahul@example.com"
}
```

The nonexistent customer should return HTTP 404. If the WCF endpoint cannot be reached or the SOAP message fails, Go maps the upstream error to HTTP 502.

## 7. Read code in request order

1. `go-gateway/cmd/api/main.go` — dependency wiring and HTTP server startup.
2. `go-gateway/internal/httpapi/handler.go` — HTTP route, validation, status mapping, JSON output.
3. `go-gateway/internal/soapclient/client.go` — SOAP/XML request, SOAPAction header, response parsing.
4. `wcf-service/Contracts/ICustomerService.cs` — the WCF contract.
5. `wcf-service/Services/CustomerService.cs` — operation implementation and validation.
6. `wcf-service/Data/CustomerRepository.cs` — SQL Server query and row mapping.
7. `database/setup.sql` — tables, seed data and stored procedure.
8. `docs/request-flow.md` — end-to-end walkthrough.
9. `docs/opensearch-nebula-notes.md` — questions to answer before adding the real search integration.

## 8. Stop and restart

Stop the Go and .NET programs with `Ctrl+C`. Stop the database but preserve data with:

```bash
docker compose down
```

Start it again later with `docker compose up -d`.

To delete the local database data permanently as well, use `docker compose down -v`. Do not do this if you need to keep the data.

## Troubleshooting

- **`dotnet: command not found`:** install `dotnet-sdk-10.0`, then open a new shell and run `dotnet --info`.
- **SQL connection refused:** verify `docker ps`, wait for SQL Server readiness, and check `docker logs learningstack-sql`.
- **SQL login fails:** confirm the same `MSSQL_SA_PASSWORD` is in `.env`, restart SQL Server if the initial container failed, then rerun database setup.
- **Database/table/procedure not found:** rerun the `sqlcmd -i /tmp/setup.sql` step and inspect any SQL errors.
- **CoreWCF can't connect to database:** ensure the `ConnectionStrings__AppDb` export is in the same terminal session where `dotnet run` is executed.
- **Go returns 502:** test `curl 'http://localhost:8081/CustomerService?wsdl'`; verify `WCF_SOAP_URL`, `SOAPAction`, operation namespace and XML response; inspect both app terminals.
- **Port busy:** find and stop the conflicting process or change the port and update `WCF_SOAP_URL`.

## OpenSearch and Nebula

This starter demonstrates the SQL Server path end to end. `docs/opensearch-nebula-notes.md` explains how to determine which component actually calls OpenSearch and where the Nebula adapter belongs. The name `Nebula` is not implemented or guessed because its internal API/protocol has not been specified.
