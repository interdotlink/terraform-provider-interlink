# terraform-provider-interlink

Terraform/OpenTofu provider for the [Inter.link](https://inter.link) Portal API.

## Build

```sh
go build ./...
```

## Regenerate the API client

The client in `internal/portal/` is generated from the Portal OpenAPI spec with
[`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen):

```sh
cd internal/portal && oapi-codegen -config config.yaml openapi.json
```
