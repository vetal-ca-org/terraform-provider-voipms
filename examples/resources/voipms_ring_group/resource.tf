data "voipms_voicemail" "sales" {
  name = "Sales"
}

data "voipms_forwarding" "mobile" {
  description = "Mobile"
}

resource "voipms_subaccount" "desk_phone" {
  username = "desk"
  password = var.desk_sip_password
}

resource "voipms_ring_group" "sales" {
  name      = "Sales"
  voicemail = data.voipms_voicemail.sales.id

  members = [
    { route = voipms_subaccount.desk_phone.route },
    { route = data.voipms_forwarding.mobile.route, ring_time = 25, press1 = true },
  ]
}

# Ring the group, and fall back to the mailbox when nobody picks up.
resource "voipms_did" "main" {
  did                  = "5550001001"
  routing              = voipms_ring_group.sales.route
  failover_noanswer    = data.voipms_voicemail.sales.route
  failover_busy        = data.voipms_voicemail.sales.route
  failover_unreachable = data.voipms_voicemail.sales.route
}
