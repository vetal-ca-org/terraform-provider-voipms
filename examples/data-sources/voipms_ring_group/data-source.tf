data "voipms_ring_group" "sales" {
  name = "Sales"
}

output "sales_route" {
  value = data.voipms_ring_group.sales.route
}
