data "voipms_ring_group" "sales" {
  name = "Sales"
}

data "voipms_voicemail" "sales" {
  name = "Sales"
}

# Ring the desks Monday to Friday 08:00-16:00 and Saturday morning; every other
# hour of the week goes straight to the mailbox.
resource "voipms_time_condition" "office_hours" {
  name            = "Office Hours"
  routing_match   = data.voipms_ring_group.sales.route
  routing_nomatch = data.voipms_voicemail.sales.route

  windows = [
    { start_day = "mon", end_day = "fri", start_hour = 8, end_hour = 16 },
    { start_day = "sat", end_day = "sat", start_hour = 9, start_minute = 30, end_hour = 12 },
  ]
}

resource "voipms_did" "main" {
  did     = "5550001001"
  routing = voipms_time_condition.office_hours.route
}
