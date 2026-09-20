---
page_title: "VoIP.ms Provider"
description: |-
  Manage VoIP.ms account objects through the REST/JSON API.
---

# VoIP.ms Provider

Use this provider to read and manage objects in a [VoIP.ms](https://voip.ms) account through the REST/JSON API.

Manage [VoIP.ms](https://voip.ms) accounts through the REST/JSON API. Authenticate with an API username (account email) and a dedicated API password from the SOAP & REST/JSON API page. The public IP of the machine running Terraform must also be allow-listed there.

This example configures a home setup: a SIP gateway (PBX trunk), two voicemail boxes, a mobile forwarding and callback, a couple of phonebook contacts, a POP lookup by hostname, and a DID that rings the gateway then fails over to voicemail or the mobile. Routes and mailbox links use computed `id` / `route` from the resource (or a data source looked up by name). Do not paste raw API ids or build `vm:` / `fwd:` strings from display names. The DID must already exist on the account; Terraform does not order or cancel numbers.

```terraform
terraform {
  required_providers {
    voipms = {
      source  = "vetal-ca-org/voipms"
      version = "~> 0.2.2"
    }
  }
}

provider "voipms" {
  # Set VOIPMS_USERNAME and VOIPMS_PASSWORD in the environment.
  # Allow-list this machine's public IP under SOAP & REST/JSON API.
}

variable "gateway_sip_password" {
  type        = string
  sensitive   = true
  description = "SIP password for the FreeSWITCH / PBX sub-account."
}

variable "voicemail_pin" {
  type        = string
  sensitive   = true
  description = "Shared mailbox PIN for the example voicemail boxes."
}

# DID routes come from resource (or data source) `route` attributes.
# Those values are API type:id strings; do not build them from display names
# and do not paste raw forwarding/mailbox ids into this file.
locals {
  route_gateway     = voipms_subaccount.gateway.route
  route_vm_alex     = voipms_voicemail.alex.route
  route_alex_mobile = voipms_forwarding.alex_mobile.route
}

# POP looked up by hostname; the data source is the source of the id.
data "voipms_server" "nyc" {
  hostname = "newyork7.voip.ms"
}

# SIP trunk for a local PBX. username is the suffix only (max 12 chars);
# the full login is account (e.g. 100001_gateway).
resource "voipms_subaccount" "gateway" {
  username              = "gateway"
  password              = var.gateway_sip_password
  description           = "Home FreeSWITCH gateway"
  device_type           = "ip_pbx"
  allowed_codecs        = "ulaw;g722"
  nat                   = "no"
  encrypted_sip_traffic = true
  canada_routing        = "premium"
}

# mailbox is the extension people dial (you choose it). id is computed and
# equals mailbox — that is what other resources reference.
resource "voipms_voicemail" "alex" {
  mailbox        = "101"
  name           = "Alex"
  password       = var.voicemail_pin
  skip_password  = true
  email          = "alex@example.com"
  attach_message = true
  timezone       = "America/New_York"
}

resource "voipms_voicemail" "jordan" {
  mailbox        = "102"
  name           = "Jordan"
  password       = var.voicemail_pin
  skip_password  = true
  email          = "jordan@example.com"
  attach_message = true
  timezone       = "America/New_York"
}

resource "voipms_forwarding" "alex_mobile" {
  phone_number = "5550002001"
  description  = "Alex mobile"
}

resource "voipms_callback" "alex_mobile" {
  description     = "Alex mobile"
  number          = "15550002001"
  callerid_number = "5550001001"
}

resource "voipms_phonebook_group" "family" {
  name = "Family"
}

resource "voipms_phonebook_entry" "alex_mobile" {
  name       = "Alex mobile"
  number     = "5550002001"
  group      = voipms_phonebook_group.family.id
  speed_dial = "21"
}

resource "voipms_phonebook_entry" "jordan_mobile" {
  name       = "Jordan mobile"
  number     = "5550002002"
  group      = voipms_phonebook_group.family.id
  speed_dial = "22"
}

# DID must already be on the account. Terraform configures routing; it does
# not order or cancel the number. Destroy only drops state.
resource "voipms_did" "home" {
  did            = "5550001001"
  note           = "Home line"
  routing      = local.route_gateway
  pop_hostname = data.voipms_server.nyc.hostname
  dialtime     = 30
  cnam         = true
  voicemail    = voipms_voicemail.alex.id

  failover_busy        = local.route_vm_alex
  failover_noanswer    = local.route_vm_alex
  failover_unreachable = local.route_alex_mobile

  sms_enabled     = true
  webhook         = "https://example.com/sms"
  webhook_enabled = true
}

resource "voipms_caller_id_filter" "blocked_prefix" {
  callerid = "999XXXXXXXX"
  did      = "all"
  routing  = "sys:hangup"
  note     = "Blocked caller ID prefix"
}
```

## Authentication

Set credentials in the provider block or with environment variables:

- `VOIPMS_USERNAME` or `voip_ms_username` — account email (API username)
- `VOIPMS_PASSWORD` or `voip_ms_api_key` — API password from **SOAP & REST/JSON API** (not the portal password)
- `VOIPMS_API_URL` — optional endpoint override

The public IP of the machine running Terraform must be allow-listed on that same portal page.

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `api_url` (String) Override the REST endpoint. Defaults to `https://voip.ms/api/v1/rest.php`. May also be set via `VOIPMS_API_URL`.
- `password` (String, Sensitive) VoIP.ms API password (not the portal password). May also be set via `VOIPMS_PASSWORD` or `voip_ms_api_key`.
- `username` (String) VoIP.ms API username (account email). May also be set via `VOIPMS_USERNAME` or `voip_ms_username`.
