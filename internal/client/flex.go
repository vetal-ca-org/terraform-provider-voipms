package client

import (
	"encoding/json"
	"strconv"
	"strings"
)

// FlexString unmarshals JSON strings, numbers, or booleans into a string.
// VoIP.ms mixes "1", 1, and 0 in the same fields across objects.
type FlexString string

func (s FlexString) String() string {
	return string(s)
}

func (s *FlexString) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*s = ""
		return nil
	}
	if len(b) > 0 && b[0] == '"' {
		var v string
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}
		*s = FlexString(v)
		return nil
	}
	if string(b) == "true" {
		*s = "1"
		return nil
	}
	if string(b) == "false" {
		*s = "0"
		return nil
	}
	*s = FlexString(strings.TrimSpace(string(b)))
	return nil
}

// Bool treats 1, true, yes, and y as true (case-insensitive).
func (s FlexString) Bool() bool {
	switch strings.ToLower(strings.TrimSpace(string(s))) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}

// Int64 parses the value as a base-10 integer. Empty values yield 0, ok=false.
func (s FlexString) Int64() (int64, bool) {
	v := strings.TrimSpace(string(s))
	if v == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}
