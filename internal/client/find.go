package client

import "strconv"

func newestMatching[T any](items []T, match func(T) bool, id func(T) string) *T {
	var best *T
	bestID := int64(-1)
	for i := range items {
		if !match(items[i]) {
			continue
		}
		n, err := strconv.ParseInt(id(items[i]), 10, 64)
		if err != nil {
			n = 0
		}
		if best == nil || n >= bestID {
			item := items[i]
			best = &item
			bestID = n
		}
	}
	return best
}

// FindForwardingAfterCreate locates a forwarding just created by phone number.
func FindForwardingAfterCreate(items []Forwarding, phone, description string) *Forwarding {
	return newestMatching(items, func(f Forwarding) bool {
		if f.PhoneNumber.String() != phone {
			return false
		}
		if description != "" && f.Description.String() != description {
			return false
		}
		return true
	}, func(f Forwarding) string { return f.Forwarding.String() })
}

// FindCallbackAfterCreate locates a callback just created by destination number.
func FindCallbackAfterCreate(items []Callback, number, description string) *Callback {
	return newestMatching(items, func(c Callback) bool {
		if c.Number.String() != number {
			return false
		}
		if description != "" && c.Description.String() != description {
			return false
		}
		return true
	}, func(c Callback) string { return c.Callback.String() })
}

// FindCallerIDFilterAfterCreate locates a filter just created by caller ID pattern.
func FindCallerIDFilterAfterCreate(items []CallerIDFilter, callerid, note string) *CallerIDFilter {
	return newestMatching(items, func(f CallerIDFilter) bool {
		if f.CallerID.String() != callerid {
			return false
		}
		if note != "" && f.Note.String() != note {
			return false
		}
		return true
	}, func(f CallerIDFilter) string { return f.Filtering.String() })
}

// FindPhonebookAfterCreate locates an entry just created by number and name.
func FindPhonebookAfterCreate(items []PhonebookEntry, number, name string) *PhonebookEntry {
	return newestMatching(items, func(p PhonebookEntry) bool {
		if p.Number.String() != number {
			return false
		}
		if name != "" && p.Name.String() != name {
			return false
		}
		return true
	}, func(p PhonebookEntry) string { return p.Phonebook.String() })
}

// FindPhonebookGroupAfterCreate locates a group just created by name.
func FindPhonebookGroupAfterCreate(items []PhonebookGroup, name string) *PhonebookGroup {
	return newestMatching(items, func(g PhonebookGroup) bool {
		return g.Name.String() == name
	}, func(g PhonebookGroup) string { return g.PhonebookGroup.String() })
}
