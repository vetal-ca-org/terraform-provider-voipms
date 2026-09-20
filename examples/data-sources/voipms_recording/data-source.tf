data "voipms_recording" "main_greeting" {
  description = "Main Greeting"
}

# Recordings are uploaded in the portal, then referenced by id. Looking one up
# by description keeps the number out of the configuration.
resource "voipms_voicemail" "main" {
  mailbox                       = "1001"
  name                          = "Main"
  password                      = var.voicemail_pin
  unavailable_message_recording = data.voipms_recording.main_greeting.id
}
