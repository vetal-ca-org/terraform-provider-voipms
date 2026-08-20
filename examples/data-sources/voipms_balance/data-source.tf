data "voipms_balance" "account" {}

output "current_balance" {
  value = data.voipms_balance.account.current_balance
}
