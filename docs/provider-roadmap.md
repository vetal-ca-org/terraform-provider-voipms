# Provider roadmap

Coverage of the VoIP.ms REST/JSON API in this Terraform provider.

**Status:** resources, single data sources, and list data sources for the table below are implemented. IVRs, ring groups, queues, time conditions, DISA, SIP URIs, recordings, call hunting, conferences, and reseller clients are not implemented.

## Implemented

| Portal area | VoIP.ms API | Terraform |
|-------------|-------------|-----------|
| Sub-accounts | `getSubAccounts` / `createSubAccount` / `setSubAccount` / `delSubAccount` | `voipms_subaccount` / `voipms_subaccounts` |
| DIDs | `getDIDsInfo` / `setDIDInfo` | `voipms_did` / `voipms_dids` (configure only; does not order or cancel numbers) |
| DID SMS / webhooks | fields on `getDIDsInfo` / `setSMS` | nested on `voipms_did` |
| Call forwarding | `getForwardings` / `createForwarding` / `setForwarding` / `delForwarding` | `voipms_forwarding` / `voipms_forwardings` |
| Voicemail boxes | `getVoicemails` / `createVoicemail` / `setVoicemail` / `delVoicemail` | `voipms_voicemail` / `voipms_voicemails` |
| Callbacks | `getCallbacks` / `createCallback` / `setCallback` / `delCallback` | `voipms_callback` / `voipms_callbacks` |
| Caller ID filtering | `getCallerIDFiltering` / `createCallerIDFiltering` / `setCallerIDFiltering` / `delCallerIDFiltering` | `voipms_caller_id_filter` / `voipms_caller_id_filters` |
| Phonebook | `getPhonebook` / `createPhonebook` / `setPhonebook` / `delPhonebook` | `voipms_phonebook_entry` / `voipms_phonebook_entries` |
| Phonebook groups | `getPhonebookGroups` / `createPhonebookGroup` / `setPhonebookGroup` / `delPhonebookGroup` | `voipms_phonebook_group` / `voipms_phonebook_groups` |
| Account balance | `getBalance` | `voipms_balance` |
| POP / servers | `getServersInfo` | `voipms_server` / `voipms_servers` |

## Not implemented

IVRs, ring groups, queues, time conditions, DISA, SIP URIs, recordings, call hunting, conferences, reseller clients, and DID order/cancel.

## Reference catalog methods (used for schema / lookup)

| API method | Purpose |
|------------|---------|
| `getServersInfo` | POP id → hostname (e.g. pop `73` → `newyork7.voip.ms`) |
| `getAuthTypes`, `getProtocols`, `getDeviceTypes` | Sub-account schema enums |
| `getAllowedCodecs`, `getDTMFModes`, `getNAT`, `getMusicOnHold`, `getLanguages` | Sub-account options |
| `getRoutes` | Canada routing (`canada_routing`: 1=Value, 2=Premium) |
| `getVoicemailSetups` | DID vs account voicemail mode |

## Example identifiers in docs

Examples use fictional values such as account `100001`, DID `5550001001`, and SIP login `100001_gateway`. Do not copy live account numbers, emails, or webhook URLs into this repository.

To inspect a private account locally, run `./scripts/dump-account.sh`. Output lands in `inventory/` and is gitignored (it includes SIP passwords).
