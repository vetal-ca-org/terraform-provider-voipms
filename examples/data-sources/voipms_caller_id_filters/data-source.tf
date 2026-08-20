data "voipms_caller_id_filters" "all" {}

output "spam_patterns" {
  value = [for f in data.voipms_caller_id_filters.all.filters : f.callerid]
}
