resource "voipms_phonebook_group" "spam" {
  name = "Spam"
}

resource "voipms_phonebook_entry" "blocked_prefix" {
  name   = "Blocked prefix"
  number = "999"
  group  = voipms_phonebook_group.spam.id
}
