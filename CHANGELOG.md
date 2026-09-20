# Changelog

All notable changes to this provider will be documented in this file.

## Unreleased

### Changed

- `voipms_subaccount` named constants: `device_type` (`ip_pbx` / `ata`), `protocol` (`sip` / `iax2`), `auth_type` (`password` / `ip`), `lock_international` (`allow` / `deny`), `international_route` and `canada_routing` (`value` / `premium`), `dialing_mode` (`main_account` / `e164` / `nanpa`), `call_pickup_behavior` (`pickup_and_be_picked_up` / `pickup_only` / `be_picked_up_only` / `disabled`). Numeric API ids still work; reads return the name.
- `voipms_did.billing_type` accepts `per_minute` / `flat` (API `1` / `2`).
- `voipms_voicemail.play_instructions` accepts `unread` / `skip_unread` (API `u` / `su`).
- `voipms_subaccount.allow225` renamed to `allow_225_balance` (state upgraded from schema v0).
- `voipms_subaccount.sip_traffic` renamed to `encrypted_sip_traffic` (state upgraded from schema v1). The VoIP.ms API field remains `sip_traffic`.
- `voipms_subaccount.pop_restriction` is unset in state when `enable_pop_restriction` is false (VoIP.ms still returns the full POP list).
- Unfiltered `getServersInfo` / `getForwardings` / `getVoicemails` / `getSubAccounts` responses are cached for the Terraform run so DID plans do not trip the VoIP.ms per-minute API limit.

### Fixed

- REST client retries Cloudflare/gateway timeouts (`522` / `523` / `524`, plus `502` / `503` / `504`) with backoff. A VoIP.ms origin blip no longer fails the whole Terraform refresh on the first HTTP error.

- `GetSubAccount` / import by numeric id: VoIP.ms `getSubAccounts?account=<id>` returns `no_subaccount`, so the client now falls back to listing all sub-accounts and matching by id.

### Added

- `voipms_did` accepts `pop_hostname` (`newyork7.voip.ms` or `New York 7`) so configs do not have to use a numeric POP id. `voipms_server` can look up a POP by `hostname` or `name` as well as `pop`.
- Named lookups: voicemail by display `name`, forwarding/callback by `description`, phonebook group/entry by `name`, sub-account by `username`. DID `voicemail_name` and phonebook `group_name` resolve the same way as `pop_hostname`.
- Resources and data sources for sub-accounts, DIDs (routing + SMS), forwarding, voicemail, callbacks, caller-ID filters, phonebook entries/groups, and POP servers.
- Single and list data sources for each of those objects.
- Resources to create/update/delete them (DID configure-only: no order/cancel).
- Registry-style docs generated with tfplugindocs.
- Provider credentials also accept `voip_ms_username` / `voip_ms_api_key`.
- Unit tests for inventory client methods and provider registration.
- `make install` / `make install-plugin` for local use from another Terraform repo.
- GitHub repository `vetal-ca-org/terraform-provider-voipms` (HashiCorp provider naming).

## 0.2.2 - 2026-08-23

### Added

- Computed `route` on `voipms_subaccount` (`account:{sip}`), `voipms_voicemail` (`vm:{mailbox}`), and `voipms_forwarding` (`fwd:{id}`), including the matching data sources.

### Changed

- DID `routing` / failover use computed `route` from a resource or data source. Attach a mailbox with `voicemail = voipms_voicemail.this.id`. Name-based `fwd:` / `vm:` strings still resolve at apply for compatibility, but configs should not paste raw API ids or display names.
