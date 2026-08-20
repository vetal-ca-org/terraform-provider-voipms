resource "voipms_caller_id_filter" "blocked_prefix" {
  callerid = "999XXXXXXXX"
  did      = "all"
  routing  = "sys:hangup"
  note     = "Blocked caller ID prefix"
}
