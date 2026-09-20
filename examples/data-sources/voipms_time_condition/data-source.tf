data "voipms_time_condition" "office_hours" {
  name = "Office Hours"
}

output "office_hours_route" {
  value = data.voipms_time_condition.office_hours.route
}
