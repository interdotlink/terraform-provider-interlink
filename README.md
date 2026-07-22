# terraform-provider-interlink

Terraform/OpenTofu provider for the [Inter.link](https://inter.link) [Portal API](https://portal.inter.link/api/v1/docs).

## Requirements

- Terraform >= 1.11 (OpenTofu >= 1.9): the `interlink_ip_transit` resource uses write-only arguments.

## Usage

```hcl
terraform {
  required_providers {
    interlink = {
      source = "interdotlink/interlink"
    }
  }
}

provider "interlink" {
  api_key = var.interlink_api_key
}
```

## Resources and data sources

- Resource `interlink_ip_transit`: manages an IP Transit service. `terraform apply` submits the order and returns immediately with the service at status `DraftQuote`; the service is **not provisioned yet**, and advances asynchronously as `status` catches up on later refreshes. Supports create, read, and import; attributes are immutable after creation, and `terraform destroy` is refused by design (cancellation is a contractual action, never automatic). Import by numeric service ID (`terraform import interlink_ip_transit.example 11`). On import the port is always reconstructed as an `existing_port` block, so configure the imported service accordingly.
- Data sources: `interlink_status`, `interlink_products`, `interlink_locations`, `interlink_projects`, `interlink_services`, and the per-family `interlink_ip_services`, `interlink_ddos_services`, `interlink_port_services`, `interlink_elan_services`, `interlink_flexpeer_services`.

See the [registry documentation](https://registry.terraform.io/providers/interdotlink/interlink/latest/docs) for full schemas and examples.

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
