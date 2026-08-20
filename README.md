# Terraform Provider for VoIP.ms

Terraform provider for [VoIP.ms](https://voip.ms). It talks to the public REST/JSON API so account objects (DIDs, subaccounts, routing, SMS, and so on) can be managed as Terraform configuration instead of clicks in the portal.

This repository follows the HashiCorp naming convention (`terraform-provider-voipms`). `main` is the default branch. See `docs/provider-roadmap.md` for API coverage.

Provider source address (local / future registry):

```hcl
source = "vetal-ca-org/voipms"
```

## What works today

Objects listed in `docs/provider-roadmap.md` are covered: a **resource** to create/update, a **data source** to read one object, and a **list data source** to read all of them.

| Kind | Name | VoIP.ms methods | Purpose |
|------|------|-----------------|---------|
| Data source | `voipms_balance` | `getBalance` | Account balance |
| Resource / data / list | `voipms_subaccount` / `voipms_subaccounts` | `getSubAccounts` / `createSubAccount` / `setSubAccount` / `delSubAccount` | SIP sub-accounts (FreeSWITCH trunk, softphones) |
| Resource / data / list | `voipms_did` / `voipms_dids` | `getDIDsInfo` / `setDIDInfo` / `setSMS` | DID routing, POP, failover, SMS webhooks. Does **not** order or cancel numbers |
| Resource / data / list | `voipms_forwarding` / `voipms_forwardings` | `getForwardings` / `setForwarding` / `delForwarding` | Call forwarding (`fwd:` targets) |
| Resource / data / list | `voipms_voicemail` / `voipms_voicemails` | `getVoicemails` / `createVoicemail` / `setVoicemail` / `delVoicemail` | Mailboxes |
| Resource / data / list | `voipms_callback` / `voipms_callbacks` | `getCallbacks` / `setCallback` / `delCallback` | Callbacks |
| Resource / data / list | `voipms_caller_id_filter` / `voipms_caller_id_filters` | `getCallerIDFiltering` / `setCallerIDFiltering` / `delCallerIDFiltering` | Spam / CID rules |
| Resource / data / list | `voipms_phonebook_entry` / `voipms_phonebook_entries` | `getPhonebook` / `setPhonebook` / `delPhonebook` | Phonebook entries |
| Resource / data / list | `voipms_phonebook_group` / `voipms_phonebook_groups` | `getPhonebookGroups` / `setPhonebookGroup` / `delPhonebookGroup` | Phonebook groups |
| Data / list | `voipms_server` / `voipms_servers` | `getServersInfo` | POP id → hostname (e.g. `73` → `newyork7.voip.ms`) |

IVRs, ring groups, queues, time conditions, DISA, SIP URIs, recordings, call hunting, conferences, and reseller clients are not implemented.

## How it fits together

```
Terraform CLI
      │  gRPC (Terraform Plugin Protocol v6)
      ▼
main.go  →  providerserver.Serve  (address: registry.terraform.io/vetal-ca-org/voipms)
      │
      ▼
internal/provider
      │  schema, Configure(), data sources, resources
      ▼
internal/client
      │  HTTP GET wrapper
      ▼
https://voip.ms/api/v1/rest.php?api_username=…&api_password=…&method=…
```

- **Plugin Framework** ([terraform-plugin-framework](https://developer.hashicorp.com/terraform/plugin/framework)) is the SDK. It is the current HashiCorp replacement for SDKv2.
- **Provider configure** reads `username` / `password` / `api_url` from the Terraform block, falling back to `VOIPMS_USERNAME` / `voip_ms_username`, `VOIPMS_PASSWORD` / `voip_ms_api_key`, and `VOIPMS_API_URL`.
- **Client** always uses GET with query parameters. VoIP.ms has no separate auth endpoint and no request body. A JSON envelope with `"status": "success"` is required; anything else becomes an `APIError`.
- **IP allow-list**: VoIP.ms rejects API calls unless the source IP is listed under **Main Menu → SOAP & REST/JSON API**. Generate an **API password** there as well; it is not the portal login password. `getIP` is the one method that works without an allow-listed IP (useful when you need to discover the address to add).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://go.dev/dl/) >= 1.25.8 (to build from source)

## Build and install locally

```shell
git clone <this-repo>
cd terraform-provider-voipms
make install
```

`make install` builds `terraform-provider-voipms` and copies it to `$(go env GOPATH)/bin`.

Terraform will not see a locally built provider unless you override provider installation. Copy `terraformrc.example` to `~/.terraformrc` (or set `TF_CLI_CONFIG_FILE`) and point the path at that `bin` directory:

```hcl
provider_installation {
  dev_overrides {
    "vetal-ca-org/voipms" = "/home/YOU/go/bin"
  }
  direct {}
}
```

`dev_overrides` skips `terraform init` downloads for that source. Use it only while developing.

## Use it from another repo (today)

Terraform providers are **binaries**, not Go modules. The other repo cannot `source = "git::..."` the way a Terraform module can. Until this is on the public Terraform Registry, pick one of these.

### Same machine (fastest): `dev_overrides`

In this provider repo:

```shell
make install
```

On the machine that runs Terraform (can be shared via `~/.terraformrc`):

```hcl
provider_installation {
  dev_overrides {
    "vetal-ca-org/voipms" = "/home/YOU/go/bin"
  }
  direct {}
}
```

In the other repo:

```hcl
terraform {
  required_providers {
    voipms = {
      source = "vetal-ca-org/voipms"
    }
  }
}

provider "voipms" {}

data "voipms_dids" "all" {}
```

Set `VOIPMS_USERNAME` and `VOIPMS_PASSWORD` in the environment. Allow-list the public IP of that machine in the VoIP.ms portal. With `dev_overrides`, skip `terraform init` for this provider; `terraform plan` talks to the local binary.

Rebuild after provider changes (`make install`) before the next plan.

### Versioned, no Registry: filesystem mirror

Installs a versioned plugin Terraform **will** download via `terraform init`:

```shell
make install-plugin          # version 0.0.1-dev by default
# VERSION=0.1.0 make install-plugin
```

That writes:

`~/.terraform.d/plugins/registry.terraform.io/vetal-ca-org/voipms/0.0.1-dev/<os>_<arch>/terraform-provider-voipms_v0.0.1-dev`

In the other repo pin the version:

```hcl
terraform {
  required_providers {
    voipms = {
      source  = "vetal-ca-org/voipms"
      version = "0.0.1-dev"
    }
  }
}
```

Then `terraform init` and `terraform plan`. Copy the same plugin directory onto any other machine that should run that config.

## Example

```hcl
terraform {
  required_providers {
    voipms = {
      source = "vetal-ca-org/voipms"
    }
  }
}

provider "voipms" {
  # Prefer environment variables in real use:
  #   export VOIPMS_USERNAME="you@example.com"
  #   export VOIPMS_PASSWORD="api-password-from-portal"
}

data "voipms_balance" "account" {}

output "balance" {
  value = data.voipms_balance.account.current_balance
}
```

Read the whole inventory (DIDs, sub-accounts, filters, …) with list data sources:

```hcl
data "voipms_dids" "all" {}
data "voipms_subaccounts" "all" {}
data "voipms_caller_id_filters" "all" {}
```

Manage an existing DID (import it first; destroy will not cancel the number):

```hcl
resource "voipms_did" "office" {
  did     = "5550001002"
  routing = "fwd:${voipms_forwarding.mobile.id}"
  pop     = 73
}

resource "voipms_forwarding" "mobile" {
  phone_number = "5550002001"
  description  = "Mobile"
}
```

More complete snippets live under [`examples/`](examples/).

## Project layout

```
main.go                 Provider process entrypoint
internal/client/        VoIP.ms REST client (no Terraform types)
internal/provider/      Framework provider, data sources, tests
examples/               Terraform snippets used by docs generation
docs/                   Registry-style documentation
.github/workflows/      CI (build + unit tests)
```

Keep API transport in `internal/client` and Terraform schema/state in `internal/provider`. That split makes the client unit-testable with `net/http/httptest` and keeps resources thin.

## Adding a resource or data source

1. Add a typed method on `internal/client` (for example `CreateSubAccount`) that calls `Client.Call`.
2. Add `*_resource.go` or `*_data_source.go` in `internal/provider` implementing the Framework interfaces (`Metadata`, `Schema`, `Configure`, plus CRUD or `Read`).
3. Register the constructor in `Resources()` or `DataSources()`.
4. Drop an example under `examples/` and a matching page under `docs/`.
5. Add an acceptance test gated on `TF_ACC=1` and real credentials.

Likely first resources, based on the VoIP.ms API surface:

- Subaccounts, DID routing/SMS, forwarding, voicemail, callbacks, caller-ID filters, and phonebook — **implemented**. See the table above.
- POP / server lookup (`getServersInfo`) as `voipms_server` / `voipms_servers` — **implemented**.

## Tests

Unit tests mock the VoIP.ms HTTP API with `net/http/httptest`. They do not need credentials or a live account.

```shell
make test
```

That covers the REST client (list/get/create/delete, empty `no_*` responses, mixed JSON string/number fields) and checks that the provider registers the inventory resources and data sources.

Acceptance tests exercise Terraform against the **live** API (`internal/provider/*_test.go` with `TestAcc…`). They skip unless credentials are set:

```shell
export TF_ACC=1
export VOIPMS_USERNAME="you@example.com"
export VOIPMS_PASSWORD="api-password"
make testacc
```

The runner’s public IP must be allow-listed. These tests read or change real account data.

CI (`.github/workflows/test.yml`) runs `go build` and `make test` only.

## Documentation

Hand-written pages live in `docs/`. After examples change, regenerate with:

```shell
make generate
```

That runs [`tfplugindocs`](https://github.com/hashicorp/terraform-plugin-docs).

## Publishing (Terraform Registry)

A Git URL is not a valid provider source. For `terraform init` in other repos without a local plugin directory, publish signed GitHub release zips and register the provider.

The GitHub repository is **`vetal-ca-org/terraform-provider-voipms`**. Terraform Registry source address is `vetal-ca-org/voipms` (the `terraform-provider-` prefix is stripped).

To publish a version:

1. [Generate a GPG key](https://developer.hashicorp.com/terraform/registry/providers/publishing#gpg-key) and add GitHub Actions secrets `GPG_PRIVATE_KEY` and `PASSPHRASE`.
2. Tag a release: `git tag v0.1.0 && git push origin v0.1.0`. `.github/workflows/release.yml` runs GoReleaser (zip + SHA256SUMS + GPG signature).
3. The repository must be **public** for the [Terraform Registry](https://registry.terraform.io/publish/provider). Sign in with GitHub and publish `vetal-ca-org/terraform-provider-voipms`.

After that, the other repo only needs:

```hcl
terraform {
  required_providers {
    voipms = {
      source  = "vetal-ca-org/voipms"
      version = "~> 0.1"
    }
  }
}
```

`terraform init` downloads the binary. No `dev_overrides` required.

Until the Registry listing exists, use `make install-plugin` or `dev_overrides` as above.

## License

Mozilla Public License 2.0. See [LICENSE](LICENSE).
