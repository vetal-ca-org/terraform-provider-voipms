package client

import "strings"

// NamedCode maps a VoIP.ms API code to a Terraform name. Numeric ids still
// round-trip; reads return the name.
type NamedCode struct {
	pairs []codePair
}

type codePair struct {
	id, name string
}

func codes(pairs ...[2]string) NamedCode {
	out := NamedCode{pairs: make([]codePair, len(pairs))}
	for i, p := range pairs {
		out.pairs[i] = codePair{id: p[0], name: p[1]}
	}
	return out
}

// Name maps an API id or name to the canonical Terraform name.
func (c NamedCode) Name(v string) (string, bool) {
	key := strings.ToLower(strings.TrimSpace(v))
	for _, p := range c.pairs {
		if key == p.id || key == p.name {
			return p.name, true
		}
	}
	return "", false
}

// ID maps an API id or name to the API code.
func (c NamedCode) ID(v string) (string, bool) {
	key := strings.ToLower(strings.TrimSpace(v))
	for _, p := range c.pairs {
		if key == p.id || key == p.name {
			return p.id, true
		}
	}
	return "", false
}

// Equal is true when a and b are the same code (name or id).
func (c NamedCode) Equal(a, b string) bool {
	ia, oka := c.ID(a)
	ib, okb := c.ID(b)
	return oka && okb && ia == ib
}

// Canada routing / international route from getRoutes: 1 = Value, 2 = Premium.
const (
	CanadaRouteValue   = "value"
	CanadaRoutePremium = "premium"
)

// CanadaRoutes is getRoutes (canada_routing and international_route).
var CanadaRoutes = codes(
	[2]string{"1", CanadaRouteValue},
	[2]string{"2", CanadaRoutePremium},
)

// CanadaRouteName maps an API id or name to the canonical name (`value` / `premium`).
func CanadaRouteName(v string) (string, bool) { return CanadaRoutes.Name(v) }

// CanadaRouteID maps an API id or name to the API id (`1` / `2`).
func CanadaRouteID(v string) (string, bool) { return CanadaRoutes.ID(v) }

// CanadaRoutesEqual is true when a and b are the same Canada route (name or id).
func CanadaRoutesEqual(a, b string) bool { return CanadaRoutes.Equal(a, b) }

// Device type from getDeviceTypes.
const (
	DeviceTypeIPPBX = "ip_pbx" // Asterisk, IP PBX, Gateway or VoIP Switch
	DeviceTypeATA   = "ata"    // ATA device, IP Phone or Softphone
)

// DeviceType is getDeviceTypes.
var DeviceType = codes(
	[2]string{"1", DeviceTypeIPPBX},
	[2]string{"2", DeviceTypeATA},
)

// Auth type from getAuthTypes.
const (
	AuthTypePassword = "password"
	AuthTypeIP       = "ip"
)

// AuthType is getAuthTypes.
var AuthType = codes(
	[2]string{"1", AuthTypePassword},
	[2]string{"2", AuthTypeIP},
)

// Protocol from getProtocols.
const (
	ProtocolSIP  = "sip"
	ProtocolIAX2 = "iax2"
)

// Protocol is getProtocols.
var Protocol = codes(
	[2]string{"1", ProtocolSIP},
	[2]string{"3", ProtocolIAX2},
)

// International calling from getLockInternational (0 = allowed, 1 = denied).
const (
	LockInternationalAllow = "allow"
	LockInternationalDeny  = "deny"
)

// LockInternational is getLockInternational.
var LockInternational = codes(
	[2]string{"0", LockInternationalAllow},
	[2]string{"1", LockInternationalDeny},
)

// Dialing mode for sub-account outbound calls.
const (
	DialingModeMainAccount = "main_account"
	DialingModeE164        = "e164"
	DialingModeNANPA       = "nanpa"
)

// DialingMode is sub-account dialing_mode.
var DialingMode = codes(
	[2]string{"0", DialingModeMainAccount},
	[2]string{"1", DialingModeE164},
	[2]string{"2", DialingModeNANPA},
)

// Call pickup behavior for sub-accounts.
const (
	CallPickupBoth       = "pickup_and_be_picked_up"
	CallPickupPickupOnly = "pickup_only"
	CallPickupBePickedUp = "be_picked_up_only"
	CallPickupDisabled   = "disabled"
)

// CallPickupBehavior is sub-account call_pickup_behavior.
var CallPickupBehavior = codes(
	[2]string{"1", CallPickupBoth},
	[2]string{"2", CallPickupPickupOnly},
	[2]string{"3", CallPickupBePickedUp},
	[2]string{"4", CallPickupDisabled},
)

// DID billing type.
const (
	BillingTypePerMinute = "per_minute"
	BillingTypeFlat      = "flat"
)

// BillingType is DID billing_type.
var BillingType = codes(
	[2]string{"1", BillingTypePerMinute},
	[2]string{"2", BillingTypeFlat},
)

// Voicemail play_instructions.
const (
	PlayInstructionsUnread     = "unread"
	PlayInstructionsSkipUnread = "skip_unread"
)

// PlayInstructions is voicemail play_instructions.
var PlayInstructions = codes(
	[2]string{"u", PlayInstructionsUnread},
	[2]string{"su", PlayInstructionsSkipUnread},
)
