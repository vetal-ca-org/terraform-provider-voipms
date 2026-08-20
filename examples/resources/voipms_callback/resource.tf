resource "voipms_callback" "mobile" {
  description      = "Mobile"
  number           = "15550002001"
  callerid_number  = "5550001002"
  delay_before     = 5
  response_timeout = 10
  digit_timeout    = 5
}
