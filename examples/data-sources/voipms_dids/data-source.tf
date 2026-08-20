data "voipms_dids" "all" {}

output "did_routing" {
  value = { for d in data.voipms_dids.all.dids : d.did => d.routing }
}
