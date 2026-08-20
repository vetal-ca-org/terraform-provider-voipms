data "voipms_subaccounts" "all" {}

output "sip_logins" {
  value = [for a in data.voipms_subaccounts.all.subaccounts : a.account]
}
