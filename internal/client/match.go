package client

import (
	"fmt"
	"strings"
)

func matchUnique[T any](items []T, query, kind string, keys func(*T) []string) (*T, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, fmt.Errorf("%w: empty %s query", ErrNotFound, kind)
	}
	var matches []*T
	for i := range items {
		item := &items[i]
		for _, key := range keys(item) {
			if key != "" && strings.EqualFold(key, q) {
				matches = append(matches, item)
				break
			}
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("%w: %s %q", ErrNotFound, kind, query)
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("%s %q matches more than one object", kind, query)
	}
	return matches[0], nil
}

func keysWithSlug(keys ...string) []string {
	out := make([]string, 0, len(keys)*2)
	seen := map[string]struct{}{}
	add := func(k string) {
		if k == "" {
			return
		}
		if _, ok := seen[k]; ok {
			return
		}
		seen[k] = struct{}{}
		out = append(out, k)
	}
	for _, k := range keys {
		add(k)
		add(Slug(k))
	}
	return out
}

// MatchVoicemail finds a mailbox by number, display name, or name slug (`main`).
func MatchVoicemail(items []Voicemail, query string) (*Voicemail, error) {
	return matchUnique(items, query, "voicemail", func(v *Voicemail) []string {
		return keysWithSlug(v.Mailbox.String(), v.Name.String())
	})
}

// MatchForwarding finds a forwarding by id, description, or description slug (`kate-fizz-cell`).
func MatchForwarding(items []Forwarding, query string) (*Forwarding, error) {
	return matchUnique(items, query, "forwarding", func(f *Forwarding) []string {
		return keysWithSlug(f.Forwarding.String(), f.Description.String())
	})
}

// MatchRingGroup finds a ring group by id, name, or name slug (`sales-line`).
func MatchRingGroup(items []RingGroup, query string) (*RingGroup, error) {
	return matchUnique(items, query, "ring group", func(g *RingGroup) []string {
		return keysWithSlug(g.RingGroup.String(), g.Name.String())
	})
}

// MatchTimeCondition finds a time condition by id, name, or name slug (`office-hours`).
func MatchTimeCondition(items []TimeCondition, query string) (*TimeCondition, error) {
	return matchUnique(items, query, "time condition", func(t *TimeCondition) []string {
		return keysWithSlug(t.TimeCondition.String(), t.Name.String())
	})
}

// MatchRecording finds a recording by id, description, or description slug
// (`main-greeting`).
func MatchRecording(items []Recording, query string) (*Recording, error) {
	return matchUnique(items, query, "recording", func(r *Recording) []string {
		return keysWithSlug(r.Recording.String(), r.Description.String())
	})
}

// MatchCallback finds a callback by id or description.
func MatchCallback(items []Callback, query string) (*Callback, error) {
	return matchUnique(items, query, "callback", func(c *Callback) []string {
		return keysWithSlug(c.Callback.String(), c.Description.String())
	})
}

// MatchPhonebookGroup finds a group by id or name.
func MatchPhonebookGroup(items []PhonebookGroup, query string) (*PhonebookGroup, error) {
	return matchUnique(items, query, "phonebook group", func(g *PhonebookGroup) []string {
		return keysWithSlug(g.PhonebookGroup.String(), g.Name.String())
	})
}

// MatchPhonebookEntry finds an entry by id or contact name.
func MatchPhonebookEntry(items []PhonebookEntry, query string) (*PhonebookEntry, error) {
	return matchUnique(items, query, "phonebook entry", func(p *PhonebookEntry) []string {
		return keysWithSlug(p.Phonebook.String(), p.Name.String())
	})
}

// NameForMailbox returns the display name for a mailbox number.
func NameForMailbox(items []Voicemail, mailbox string) string {
	for i := range items {
		if items[i].Mailbox.String() == mailbox {
			return items[i].Name.String()
		}
	}
	return ""
}
