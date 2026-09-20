package client

import (
	"fmt"
	"strings"
)

// VoIP.ms DID routing prefixes. The target after the colon is the API id
// (mailbox, forwarding id) or the SIP login for account: routes.
const (
	RouteKindAccount  = "account"
	RouteKindFwd      = "fwd"
	RouteKindVM       = "vm"
	RouteKindGroup    = "grp"
	RouteKindTimeCond = "tc"
)

// AccountRoute is the DID routing value for a sub-account SIP login.
func AccountRoute(account string) string { return RouteKindAccount + ":" + account }

// ForwardingRoute is the DID routing value for a forwarding id.
func ForwardingRoute(id string) string { return RouteKindFwd + ":" + id }

// VoicemailRoute is the DID routing value for a mailbox id.
func VoicemailRoute(mailbox string) string { return RouteKindVM + ":" + mailbox }

// RingGroupRoute is the DID routing value for a ring group id.
func RingGroupRoute(id string) string { return RouteKindGroup + ":" + id }

// TimeConditionRoute is the DID routing value for a time condition id.
func TimeConditionRoute(id string) string { return RouteKindTimeCond + ":" + id }

// RouteTables is the account data needed to resolve named DID routes.
type RouteTables struct {
	Forwardings    []Forwarding
	Voicemails     []Voicemail
	RingGroups     []RingGroup
	TimeConditions []TimeCondition
}

// CanonicalRoute rewrites fwd:/vm:/grp:/tc: targets to the numeric API form.
// Other prefixes (account:, sys:, none:) are returned unchanged.
func CanonicalRoute(route string, tables RouteTables) (string, error) {
	route = strings.TrimSpace(route)
	kind, rest, ok := strings.Cut(route, ":")
	if !ok {
		return route, nil
	}
	rest = strings.TrimSpace(rest)
	switch strings.ToLower(kind) {
	case "fwd":
		fwd, err := MatchForwarding(tables.Forwardings, rest)
		if err != nil {
			return "", fmt.Errorf("route %q: %w", route, err)
		}
		return "fwd:" + fwd.Forwarding.String(), nil
	case "vm":
		if rest == "" || rest == "0" {
			return route, nil
		}
		box, err := MatchVoicemail(tables.Voicemails, rest)
		if err != nil {
			return "", fmt.Errorf("route %q: %w", route, err)
		}
		return "vm:" + box.Mailbox.String(), nil
	case "grp":
		group, err := MatchRingGroup(tables.RingGroups, rest)
		if err != nil {
			return "", fmt.Errorf("route %q: %w", route, err)
		}
		return "grp:" + group.RingGroup.String(), nil
	case "tc":
		cond, err := MatchTimeCondition(tables.TimeConditions, rest)
		if err != nil {
			return "", fmt.Errorf("route %q: %w", route, err)
		}
		return "tc:" + cond.TimeCondition.String(), nil
	default:
		return route, nil
	}
}

// RoutesEqual is true when a and b resolve to the same API route.
func RoutesEqual(a, b string, tables RouteTables) bool {
	ca, errA := CanonicalRoute(a, tables)
	cb, errB := CanonicalRoute(b, tables)
	return errA == nil && errB == nil && ca == cb
}
