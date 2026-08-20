# Changelog

All notable changes to this provider will be documented in this file.

## Unreleased

### Added

- Resources and data sources for sub-accounts, DIDs (routing + SMS), forwarding, voicemail, callbacks, caller-ID filters, phonebook entries/groups, and POP servers.
- Single and list data sources for each of those objects.
- Resources to create/update/delete them (DID configure-only: no order/cancel).
- Registry-style docs generated with tfplugindocs.
- Provider credentials also accept `voip_ms_username` / `voip_ms_api_key`.
- Unit tests for inventory client methods and provider registration.
- `make install` / `make install-plugin` for local use from another Terraform repo.
- GitHub repository `vetal-ca-org/terraform-provider-voipms` (HashiCorp provider naming).
