data "voipms_server" "newyork7" {
  pop = "73"
}

output "hostname" {
  value = data.voipms_server.newyork7.hostname
}
