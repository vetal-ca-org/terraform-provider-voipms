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
