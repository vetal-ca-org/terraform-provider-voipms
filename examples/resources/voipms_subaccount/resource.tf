resource "voipms_subaccount" "gateway" {
  username              = "gateway"
  password              = var.gateway_sip_password
  description           = "Common SIP gateway"
  device_type           = "ip_pbx"
  allowed_codecs        = "ulaw;g722"
  nat                   = "no"
  encrypted_sip_traffic = true
  canada_routing        = "premium"
  allow_225_balance     = false
}
