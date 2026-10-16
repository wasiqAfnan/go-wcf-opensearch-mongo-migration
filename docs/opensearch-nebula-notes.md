# Adding the OpenSearch / Nebula path

The meaning of `Nebula` has not been specified in this learning lab. It could be a company-internal SDK, gateway, proxy, query builder, service, or a separately named product. Do not assume it is NebulaGraph or that it communicates with OpenSearch over a particular protocol until you inspect the real code/configuration.

First answer these questions in your workplace repository:

1. Where is `Nebula` imported, configured, or constructed? Search by package name, namespace, configuration key, endpoint URL, and class/client name.
2. Is `Nebula` a library linked into the Go process, a library linked into WCF/.NET, or a separate network service?
3. Which component sends the actual request to OpenSearch?
4. Where are the OpenSearch URL, authentication, index/alias, query DSL, timeout and TLS settings configured?
5. Is search invoked directly from Go, through a WCF operation, or through a separate backend service?
6. What is returned: raw OpenSearch hits, a mapped domain object, or an aggregated response?

## Possible flow A: Go calls WCF, and WCF uses Nebula/OpenSearch

```text
HTTP client
  -> Go API handler
  -> Go SOAP client
  -> WCF search operation
  -> Nebula adapter/client (actual implementation depends on your system)
  -> OpenSearch
  <- search hits
  <- WCF SOAP response
  <- Go JSON response
```

## Possible flow B: Go calls Nebula/OpenSearch directly

```text
HTTP client -> Go API handler -> Nebula/OpenSearch client -> OpenSearch -> Go JSON response
```

These are different architectures. Inspect the call sites to determine which one your project uses.

## Recommended design boundary

Keep search-specific work behind an interface such as `SearchProvider` or `NebulaClient`, whichever fits your codebase. The handler should validate the query and call the application service; the adapter should handle the external protocol and map its response. Avoid scattering OpenSearch query JSON across unrelated handlers.

Do not merge the search path into this starter until you confirm the actual Nebula interface. A small example based on a guessed integration would teach the wrong thing.
