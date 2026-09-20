data "voipms_time_conditions" "all" {}

output "time_condition_names" {
  value = [for c in data.voipms_time_conditions.all.time_conditions : c.name]
}
