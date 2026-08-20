#!/usr/bin/env bash
# Dump VoIP.ms account configuration via the REST API.
#
# Credentials (same names as the curl example in the portal):
#   voip_ms_username   API username (account email)
#   voip_ms_api_key      API password from SOAP & REST/JSON API
#
# Optional:
#   VOIPMS_API_URL       override REST endpoint
#   OUT                  output path (default: inventory/account-YYYYMMDD.json)
#
# The JSON includes sub-account SIP passwords. Do not commit inventory/ output.

set -euo pipefail

: "${voip_ms_username:?set voip_ms_username}"
: "${voip_ms_api_key:?set voip_ms_api_key}"

API="${VOIPMS_API_URL:-https://voip.ms/api/v1/rest.php}"
OUT="${OUT:-inventory/account-$(date +%Y%m%d).json}"
mkdir -p "$(dirname "$OUT")"

call_api() {
  curl --fail-with-body --silent --show-error --get "$API" \
    --data-urlencode "api_username=${voip_ms_username}" \
    --data-urlencode "api_password=${voip_ms_api_key}" \
    --data-urlencode "method=$1" "${@:2}"
}

entries=(
  "balance:getBalance"
  "sub_accounts:getSubAccounts"
  "dids:getDIDsInfo"
  "forwardings:getForwardings"
  "voicemails:getVoicemails"
  "callbacks:getCallbacks"
  "phonebook:getPhonebook"
  "phonebook_groups:getPhonebookGroups"
  "caller_id_filtering:getCallerIDFiltering"
  "locations:getLocations"
  "ivrs:getIVRs"
  "ring_groups:getRingGroups"
  "queues:getQueues"
  "time_conditions:getTimeConditions"
  "disas:getDISAs"
  "sip_uris:getSIPURIs"
  "recordings:getRecordings"
  "call_huntings:getCallHuntings"
  "conferences:getConference"
  "clients:getClients"
)

{
  echo '{'
  first=1
  for entry in "${entries[@]}"; do
    key="${entry%%:*}"
    method="${entry##*:}"
    body=$(call_api "$method")
    if [[ "$first" -eq 1 ]]; then first=0; else echo ','; fi
    printf '  "%s": %s' "$key" "$body"
  done
  echo
  echo '}'
} >"$OUT"

echo "Wrote $OUT ($(wc -c <"$OUT") bytes)"
echo
echo "Redacted summary:"
jq '{
  balance: .balance.balance.current_balance,
  sub_accounts: [.sub_accounts.accounts[]? | {account, username, description, device_type, allowed_codecs, nat, sip_traffic}],
  dids: [.dids.dids[]? | {did, note, routing, pop, voicemail, e911, sms_enabled, webhook_enabled, sms_email_enabled, sms_url_callback_enabled}],
  forwardings: [.forwardings.forwardings[]? | {forwarding, phone_number, description}],
  voicemails: [.voicemails.voicemails[]? | {mailbox, name, email, timezone}],
  callbacks: [.callbacks.callbacks[]? | {callback, description, number, callerid_number}],
  phonebook_groups: .phonebook_groups.phonebook_groups,
  phonebook_count: (.phonebook.phonebooks | length),
  caller_id_filter_count: (.caller_id_filtering.filtering | length),
  unconfigured: {
    ivrs: .ivrs.status,
    ring_groups: .ring_groups.status,
    queues: .queues.status,
    time_conditions: .time_conditions.status,
    disas: .disas.status,
    sip_uris: .sip_uris.status,
    recordings: .recordings.status,
    call_huntings: .call_huntings.status,
    conferences: .conferences.status,
    clients: .clients.status
  }
}' "$OUT"
