resource "voipms_subaccount" "gateway" {
  username        = "gateway"
  password        = var.gateway_sip_password
  description     = "Common SIP gateway"
  protocol        = "1"
  auth_type       = "1"
  device_type     = "1"
  allowed_codecs  = "ulaw;g722"
  nat             = "no"
  sip_traffic     = true
  canada_routing  = "2"
}
