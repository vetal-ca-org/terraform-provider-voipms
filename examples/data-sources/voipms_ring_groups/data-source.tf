data "voipms_ring_groups" "all" {}

output "ring_group_names" {
  value = [for g in data.voipms_ring_groups.all.ring_groups : g.name]
}
