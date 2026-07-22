# Contributing

Thanks for your interest in improving the Inter.link Terraform provider. Bug reports, documentation fixes, and code contributions are all welcome.

For anything bigger than a small fix, please [open an issue](https://github.com/interdotlink/terraform-provider-interlink/issues/new) first to discuss the change before you invest time in it.

## Development setup

You need Go (the version pinned in `go.mod`; the toolchain downloads it automatically if yours is older). Then:

```sh
git clone https://github.com/YOUR-FORK/terraform-provider-interlink
cd terraform-provider-interlink
go build ./...
go test ./...
```

Please run `gofmt` and `go vet ./...` before submitting.

## Generated code

Two parts of the repo are generated. Never edit them by hand; change the source and regenerate instead:

- **`internal/portal/`** is the Portal API client, generated from the OpenAPI spec (browsable at the [Portal API docs](https://portal.inter.link/api/v1/docs)):

  ```sh
  cd internal/portal && oapi-codegen -config config.yaml openapi.json
  ```

- **`docs/`** is rendered by [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs). Edit the schema descriptions in `internal/provider/` or the examples in `examples/`, then regenerate (needs the `terraform` CLI on your PATH):

  ```sh
  go generate ./...
  ```

## Testing

Unit tests need no credentials: `go test ./...`

Acceptance tests run against the real Portal API and are skipped unless `TF_ACC` is set:

```sh
TF_ACC=1 INTERLINK_API_KEY=your-key go test ./internal/provider/ -v
```

**Never point write operations at production.** Creating an `interlink_ip_transit` resource submits a real, billable order the moment the API accepts it, and it cannot be cancelled through the API. The acceptance-test suite is read-only on purpose; do not add tests that create resources.

## Pull requests

- Branch from `main` and keep each PR focused on one change.
- Add or update tests for behavior changes.
- Regenerate `docs/` if you touched a schema or an attribute description.
- Start commit messages with an imperative verb: Add, Fix, Change, Remove, Update, Refactor.
- Maintainers handle versioning and releases (a pushed tag triggers the release workflow), so please don't bump versions in your PR.

## License

The provider is licensed under [MPL-2.0](LICENSE). By contributing, you agree that your work is provided under the same license.
