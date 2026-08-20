resource "voipms_did" "home" {
  did                  = "5550001001"
  note                 = "Home line"
  routing              = "account:100001_gateway"
  pop                  = 73
  dialtime             = 30
  voicemail            = "101"
  failover_busy        = "vm:101"
  failover_noanswer    = "vm:101"
  failover_unreachable = "vm:101"

  sms_enabled     = true
  webhook         = "https://example.lambda-url.us-east-1.on.aws/"
  webhook_enabled = true
}
