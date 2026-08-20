resource "voipms_voicemail" "main" {
  mailbox        = "101"
  name           = "Main"
  password       = var.voicemail_pin
  skip_password  = true
  email          = "you@example.com"
  attach_message = true
  timezone       = "America/Montreal"
  language       = "en"
}
