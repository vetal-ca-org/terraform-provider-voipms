package client

// SubAccount is a SIP/IAX sub-account from getSubAccounts.
type SubAccount struct {
	ID                        FlexString `json:"id"`
	Account                   FlexString `json:"account"`
	Username                  FlexString `json:"username"`
	Description               FlexString `json:"description"`
	Protocol                  FlexString `json:"protocol"`
	AuthType                  FlexString `json:"auth_type"`
	Password                  FlexString `json:"password"`
	IP                        FlexString `json:"ip"`
	DeviceType                FlexString `json:"device_type"`
	CallerIDNumber            FlexString `json:"callerid_number"`
	CanadaRouting             FlexString `json:"canada_routing"`
	LockInternational         FlexString `json:"lock_international"`
	InternationalRoute        FlexString `json:"international_route"`
	MusicOnHold               FlexString `json:"music_on_hold"`
	Language                  FlexString `json:"language"`
	AllowedCodecs             FlexString `json:"allowed_codecs"`
	DTMFMode                  FlexString `json:"dtmf_mode"`
	NAT                       FlexString `json:"nat"`
	SIPTraffic                FlexString `json:"sip_traffic"`
	MaxExpiry                 FlexString `json:"max_expiry"`
	RTPTimeout                FlexString `json:"rtp_timeout"`
	RTPHoldTimeout            FlexString `json:"rtp_hold_timeout"`
	IPRestriction             FlexString `json:"ip_restriction"`
	EnableIPRestriction       FlexString `json:"enable_ip_restriction"`
	POPRestriction            FlexString `json:"pop_restriction"`
	EnablePOPRestriction      FlexString `json:"enable_pop_restriction"`
	RecordCalls               FlexString `json:"record_calls"`
	Allow225                  FlexString `json:"allow225"`
	InternalExtension         FlexString `json:"internal_extension"`
	InternalVoicemail         FlexString `json:"internal_voicemail"`
	InternalDialtime          FlexString `json:"internal_dialtime"`
	EnableInternalCNAM        FlexString `json:"enable_internal_cnam"`
	DialingMode               FlexString `json:"dialing_mode"`
	DefaultE911               FlexString `json:"default_e911"`
	CallPickupBehavior        FlexString `json:"call_pickup_behavior"`
	InternalExtensionLocation FlexString `json:"internal_extension_location"`
	ParkingLot                FlexString `json:"parking_lot"`
	TFCarrier                 FlexString `json:"tfcarrier"`
	TranscriptionStartDelay   FlexString `json:"transcription_start_delay"`
}

func (s SubAccount) SetParams() map[string]string {
	return map[string]string{
		"id":                          s.ID.String(),
		"description":                 s.Description.String(),
		"auth_type":                   s.AuthType.String(),
		"password":                    s.Password.String(),
		"ip":                          s.IP.String(),
		"device_type":                 s.DeviceType.String(),
		"callerid_number":             s.CallerIDNumber.String(),
		"canada_routing":              s.CanadaRouting.String(),
		"lock_international":          s.LockInternational.String(),
		"international_route":         s.InternationalRoute.String(),
		"music_on_hold":               s.MusicOnHold.String(),
		"language":                    s.Language.String(),
		"record_calls":                s.RecordCalls.String(),
		"allowed_codecs":              s.AllowedCodecs.String(),
		"dtmf_mode":                   s.DTMFMode.String(),
		"nat":                         s.NAT.String(),
		"sip_traffic":                 s.SIPTraffic.String(),
		"max_expiry":                  s.MaxExpiry.String(),
		"rtp_timeout":                 s.RTPTimeout.String(),
		"rtp_hold_timeout":            s.RTPHoldTimeout.String(),
		"ip_restriction":              s.IPRestriction.String(),
		"enable_ip_restriction":       s.EnableIPRestriction.String(),
		"pop_restriction":             s.POPRestriction.String(),
		"enable_pop_restriction":      s.EnablePOPRestriction.String(),
		"internal_extension":          s.InternalExtension.String(),
		"internal_voicemail":          s.InternalVoicemail.String(),
		"internal_dialtime":           s.InternalDialtime.String(),
		"allow225":                    s.Allow225.String(),
		"enable_internal_cnam":        s.EnableInternalCNAM.String(),
		"dialing_mode":                s.DialingMode.String(),
		"default_e911":                s.DefaultE911.String(),
		"call_pickup_behavior":        s.CallPickupBehavior.String(),
		"internal_extension_location": s.InternalExtensionLocation.String(),
		"parking_lot":                 s.ParkingLot.String(),
		"tfcarrier":                   s.TFCarrier.String(),
		"transcription_start_delay":   s.TranscriptionStartDelay.String(),
	}
}

// DID is a phone number from getDIDsInfo.
type DID struct {
	DID                     FlexString `json:"did"`
	Description             FlexString `json:"description"`
	VoicemailThreshold      FlexString `json:"voicemail_threshold"`
	Routing                 FlexString `json:"routing"`
	FailoverBusy            FlexString `json:"failover_busy"`
	FailoverUnreachable     FlexString `json:"failover_unreachable"`
	FailoverNoanswer        FlexString `json:"failover_noanswer"`
	Voicemail               FlexString `json:"voicemail"`
	POP                     FlexString `json:"pop"`
	Dialtime                FlexString `json:"dialtime"`
	CNAM                    FlexString `json:"cnam"`
	E911                    FlexString `json:"e911"`
	CallerIDPrefix          FlexString `json:"callerid_prefix"`
	RecordCalls             FlexString `json:"record_calls"`
	Note                    FlexString `json:"note"`
	BillingType             FlexString `json:"billing_type"`
	NextBilling             FlexString `json:"next_billing"`
	OrderDate               FlexString `json:"order_date"`
	SMSAvailable            FlexString `json:"sms_available"`
	SMSEnabled              FlexString `json:"sms_enabled"`
	MMSAvailable            FlexString `json:"mms_available"`
	SMSEmail                FlexString `json:"sms_email"`
	SMSEmailEnabled         FlexString `json:"sms_email_enabled"`
	SMSForward              FlexString `json:"sms_forward"`
	SMSForwardEnabled       FlexString `json:"sms_forward_enabled"`
	SMSURLCallback          FlexString `json:"sms_url_callback"`
	SMSURLCallbackEnabled   FlexString `json:"sms_url_callback_enabled"`
	SMSURLCallbackRetry     FlexString `json:"sms_url_callback_retry"`
	WebhookEnabled          FlexString `json:"webhook_enabled"`
	Webhook                 FlexString `json:"webhook"`
	Dialmode                FlexString `json:"dialmode"`
	SMSSIPAccount           FlexString `json:"sms_sipaccount"`
	SMSSIPAccountEnabled    FlexString `json:"sms_sipaccount_enabled"`
	InboundDialingMode      FlexString `json:"inbound_dialing_mode"`
	PortOutPIN              FlexString `json:"port_out_pin"`
	Transcribe              FlexString `json:"transcribe"`
	TranscriptionEmail      FlexString `json:"transcription_email"`
	TranscriptionLocale     FlexString `json:"transcription_locale"`
	TranscriptionRedaction  FlexString `json:"transcription_redaction"`
	TranscriptionSentiment  FlexString `json:"transcription_sentiment"`
	TranscriptionSummary    FlexString `json:"transcription_summary"`
	TranscriptionStartDelay FlexString `json:"transcription_start_delay"`
}

func (d DID) SetInfoParams() map[string]string {
	return map[string]string{
		"did":                       d.DID.String(),
		"routing":                   d.Routing.String(),
		"failover_busy":             d.FailoverBusy.String(),
		"failover_unreachable":      d.FailoverUnreachable.String(),
		"failover_noanswer":         d.FailoverNoanswer.String(),
		"voicemail":                 d.Voicemail.String(),
		"pop":                       d.POP.String(),
		"dialtime":                  d.Dialtime.String(),
		"cnam":                      d.CNAM.String(),
		"callerid_prefix":           d.CallerIDPrefix.String(),
		"note":                      d.Note.String(),
		"billing_type":              d.BillingType.String(),
		"record_calls":              d.RecordCalls.String(),
		"voicemail_threshold":       d.VoicemailThreshold.String(),
		"inbound_dialing_mode":      d.InboundDialingMode.String(),
		"port_out_pin":              d.PortOutPIN.String(),
		"transcribe":                zeroOne(d.Transcribe),
		"transcription_email":       d.TranscriptionEmail.String(),
		"transcription_locale":      d.TranscriptionLocale.String(),
		"transcription_redaction":   yesNo(d.TranscriptionRedaction),
		"transcription_sentiment":   yesNo(d.TranscriptionSentiment),
		"transcription_summary":     yesNo(d.TranscriptionSummary),
		"transcription_start_delay": d.TranscriptionStartDelay.String(),
	}
}

func (d DID) SetSMSParams() map[string]string {
	return map[string]string{
		"did":                    d.DID.String(),
		"enable":                 d.SMSEnabled.String(),
		"email_enabled":          d.SMSEmailEnabled.String(),
		"email_address":          d.SMSEmail.String(),
		"sms_forward_enable":     d.SMSForwardEnabled.String(),
		"sms_forward":            d.SMSForward.String(),
		"url_callback_enable":    d.SMSURLCallbackEnabled.String(),
		"url_callback":           d.SMSURLCallback.String(),
		"url_callback_retry":     d.SMSURLCallbackRetry.String(),
		"sms_sipaccount":         d.SMSSIPAccount.String(),
		"sms_sipaccount_enabled": d.SMSSIPAccountEnabled.String(),
		"webhook":                d.Webhook.String(),
		"webhook_enabled":        d.WebhookEnabled.String(),
	}
}

// Forwarding is a call-forward destination from getForwardings.
type Forwarding struct {
	Forwarding       FlexString `json:"forwarding"`
	PhoneNumber      FlexString `json:"phone_number"`
	CallerIDOverride FlexString `json:"callerid_override"`
	Description      FlexString `json:"description"`
	DTMFDigits       FlexString `json:"dtmf_digits"`
	Pause            FlexString `json:"pause"`
	DiversionHeader  FlexString `json:"diversion_header"`
}

func (f Forwarding) SetParams() map[string]string {
	return map[string]string{
		"forwarding":        f.Forwarding.String(),
		"phone_number":      f.PhoneNumber.String(),
		"callerid_override": f.CallerIDOverride.String(),
		"description":       f.Description.String(),
		"dtmf_digits":       f.DTMFDigits.String(),
		"pause":             f.Pause.String(),
		"diversion_header":  f.DiversionHeader.String(),
	}
}

// Voicemail is a mailbox from getVoicemails.
type Voicemail struct {
	Mailbox                     FlexString `json:"mailbox"`
	Name                        FlexString `json:"name"`
	Password                    FlexString `json:"password"`
	SkipPassword                FlexString `json:"skip_password"`
	Email                       FlexString `json:"email"`
	AttachMessage               FlexString `json:"attach_message"`
	DeleteMessage               FlexString `json:"delete_message"`
	SayTime                     FlexString `json:"say_time"`
	Timezone                    FlexString `json:"timezone"`
	SayCallerID                 FlexString `json:"say_callerid"`
	PlayInstructions            FlexString `json:"play_instructions"`
	Language                    FlexString `json:"language"`
	EmailAttachmentFormat       FlexString `json:"email_attachment_format"`
	UnavailableMessageRecording FlexString `json:"unavailable_message_recording"`
	Client                      FlexString `json:"client"`
	Transcription               FlexString `json:"transcription"`
	TranscriptionLocale         FlexString `json:"transcription_locale"`
	TranscriptionRedaction      FlexString `json:"transcription_redaction"`
	TranscriptionSummary        FlexString `json:"transcription_summary"`
	TranscriptionSentiment      FlexString `json:"transcription_sentiment"`
	TranscriptionFormat         FlexString `json:"transcription_format"`
}

func (v Voicemail) SetParams() map[string]string {
	params := map[string]string{
		"mailbox":                       v.Mailbox.String(),
		"name":                          v.Name.String(),
		"password":                      v.Password.String(),
		"skip_password":                 yesNo(v.SkipPassword),
		"email":                         v.Email.String(),
		"attach_message":                yesNo(v.AttachMessage),
		"delete_message":                yesNo(v.DeleteMessage),
		"say_time":                      yesNo(v.SayTime),
		"timezone":                      v.Timezone.String(),
		"say_callerid":                  yesNo(v.SayCallerID),
		"play_instructions":             v.PlayInstructions.String(),
		"language":                      v.Language.String(),
		"email_attachment_format":       v.EmailAttachmentFormat.String(),
		"unavailable_message_recording": v.UnavailableMessageRecording.String(),
		"transcription":                 yesNo(v.Transcription),
		"transcription_locale":          v.TranscriptionLocale.String(),
		"transcription_redaction":       yesNo(v.TranscriptionRedaction),
		"transcription_summary":         yesNo(v.TranscriptionSummary),
		"transcription_sentiment":       yesNo(v.TranscriptionSentiment),
		"transcription_format":          v.TranscriptionFormat.String(),
	}
	// Reseller-only. Sending client=0 on a non-reseller account is rejected.
	if id := v.Client.String(); id != "" && id != "0" {
		params["client"] = id
	}
	return params
}

// zeroOne renders a VoIP.ms boolean as 1/0.
func zeroOne(v FlexString) string {
	if v.Bool() {
		return "1"
	}
	return "0"
}

// yesNo renders a VoIP.ms boolean for a set* call. getVoicemails answers with
// 1/0 or Y/N depending on the field; every setVoicemail flag wants yes/no.
func yesNo(v FlexString) string {
	if v.Bool() {
		return "yes"
	}
	return "no"
}

// RingGroup is a ring group from getRingGroups.
type RingGroup struct {
	RingGroup          FlexString `json:"ring_group"`
	Name               FlexString `json:"name"`
	Members            FlexString `json:"members"`
	Voicemail          FlexString `json:"voicemail"`
	CallerAnnouncement FlexString `json:"caller_announcement"`
	MusicOnHold        FlexString `json:"music_on_hold"`
	Language           FlexString `json:"language"`
}

func (g RingGroup) SetParams() map[string]string {
	return map[string]string{
		"ring_group":          g.RingGroup.String(),
		"name":                g.Name.String(),
		"members":             g.Members.String(),
		"voicemail":           g.Voicemail.String(),
		"caller_announcement": g.CallerAnnouncement.String(),
		"music_on_hold":       g.MusicOnHold.String(),
		"language":            g.Language.String(),
	}
}

// TimeCondition is a time-of-day routing rule from getTimeConditions. The six
// window fields are parallel semicolon-separated lists — `starthour` "8;9" with
// `weekdaystart` "mon;sat" is two windows, and VoIP.ms rejects a write whose
// lists are not all the same length.
type TimeCondition struct {
	TimeCondition  FlexString `json:"timecondition"`
	Name           FlexString `json:"name"`
	RoutingMatch   FlexString `json:"routing_match"`
	RoutingNomatch FlexString `json:"routing_nomatch"`
	StartHour      FlexString `json:"starthour"`
	StartMinute    FlexString `json:"startminute"`
	EndHour        FlexString `json:"endhour"`
	EndMinute      FlexString `json:"endminute"`
	WeekdayStart   FlexString `json:"weekdaystart"`
	WeekdayEnd     FlexString `json:"weekdayend"`
}

func (t TimeCondition) SetParams() map[string]string {
	return map[string]string{
		"timecondition":   t.TimeCondition.String(),
		"name":            t.Name.String(),
		"routing_match":   t.RoutingMatch.String(),
		"routing_nomatch": t.RoutingNomatch.String(),
		"starthour":       t.StartHour.String(),
		"startminute":     t.StartMinute.String(),
		"endhour":         t.EndHour.String(),
		"endminute":       t.EndMinute.String(),
		"weekdaystart":    t.WeekdayStart.String(),
		"weekdayend":      t.WeekdayEnd.String(),
	}
}

// Recording is an audio prompt from getRecordings. VoIP.ms answers catalog
// style — value/description, not id/name — the way getMusicOnHold does.
type Recording struct {
	Recording   FlexString `json:"value"`
	Description FlexString `json:"description"`
}

// Callback is a callback from getCallbacks.
type Callback struct {
	Callback        FlexString `json:"callback"`
	Description     FlexString `json:"description"`
	Number          FlexString `json:"number"`
	DelayBefore     FlexString `json:"delay_before"`
	ResponseTimeout FlexString `json:"response_timeout"`
	DigitTimeout    FlexString `json:"digit_timeout"`
	CallerIDNumber  FlexString `json:"callerid_number"`
}

func (c Callback) SetParams() map[string]string {
	return map[string]string{
		"callback":         c.Callback.String(),
		"description":      c.Description.String(),
		"number":           c.Number.String(),
		"delay_before":     c.DelayBefore.String(),
		"response_timeout": c.ResponseTimeout.String(),
		"digit_timeout":    c.DigitTimeout.String(),
		"callerid_number":  c.CallerIDNumber.String(),
	}
}

// CallerIDFilter is a caller-ID routing rule from getCallerIDFiltering.
type CallerIDFilter struct {
	Filtering           FlexString `json:"filtering"`
	CallerID            FlexString `json:"callerid"`
	DID                 FlexString `json:"did"`
	Routing             FlexString `json:"routing"`
	FailoverUnreachable FlexString `json:"failover_unreachable"`
	FailoverBusy        FlexString `json:"failover_busy"`
	FailoverNoanswer    FlexString `json:"failover_noanswer"`
	Note                FlexString `json:"note"`
}

func (f CallerIDFilter) SetParams() map[string]string {
	return map[string]string{
		"filtering":            f.Filtering.String(),
		"callerid":             f.CallerID.String(),
		"did":                  f.DID.String(),
		"routing":              f.Routing.String(),
		"failover_unreachable": f.FailoverUnreachable.String(),
		"failover_busy":        f.FailoverBusy.String(),
		"failover_noanswer":    f.FailoverNoanswer.String(),
		"note":                 f.Note.String(),
	}
}

// PhonebookEntry is a speed-dial / phonebook row from getPhonebook.
type PhonebookEntry struct {
	Phonebook FlexString `json:"phonebook"`
	SpeedDial FlexString `json:"speed_dial"`
	Name      FlexString `json:"name"`
	Number    FlexString `json:"number"`
	CallerID  FlexString `json:"callerid"`
	Note      FlexString `json:"note"`
	Group     FlexString `json:"group"`
	GroupName FlexString `json:"group_name"`
}

func (p PhonebookEntry) SetParams() map[string]string {
	return map[string]string{
		"phonebook":  p.Phonebook.String(),
		"speed_dial": p.SpeedDial.String(),
		"name":       p.Name.String(),
		"number":     p.Number.String(),
		"callerid":   p.CallerID.String(),
		"note":       p.Note.String(),
		"group":      p.Group.String(),
	}
}

// PhonebookGroup is a phonebook group from getPhonebookGroups.
type PhonebookGroup struct {
	PhonebookGroup FlexString `json:"phonebook_group"`
	Name           FlexString `json:"name"`
	Members        FlexString `json:"members"`
}

func (g PhonebookGroup) SetParams() map[string]string {
	return map[string]string{
		"group":   g.PhonebookGroup.String(),
		"name":    g.Name.String(),
		"members": g.Members.String(),
	}
}

// Server is a VoIP.ms POP from getServersInfo.
type Server struct {
	Name        FlexString `json:"server_name"`
	Shortname   FlexString `json:"server_shortname"`
	Hostname    FlexString `json:"server_hostname"`
	IP          FlexString `json:"server_ip"`
	Country     FlexString `json:"server_country"`
	POP         FlexString `json:"server_pop"`
	Recommended FlexString `json:"server_recommended"`
}
