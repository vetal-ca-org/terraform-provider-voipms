data "voipms_servers" "all" {}

output "pop_hostnames" {
  value = { for s in data.voipms_servers.all.servers : s.pop => s.hostname }
}
