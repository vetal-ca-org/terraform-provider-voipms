data "voipms_recordings" "all" {}

output "recordings_by_description" {
  value = { for r in data.voipms_recordings.all.recordings : r.description => r.id }
}
