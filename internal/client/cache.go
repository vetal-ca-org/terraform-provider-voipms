package client

import "sync"

// listCache holds unfiltered account lists for one Terraform run so DID
// Read/ModifyPlan do not hit VoIP.ms once per object (API rate limit).
type listCache struct {
	mu           sync.Mutex
	servers      []Server
	forwardings  []Forwarding
	voicemails   []Voicemail
	subaccounts  []SubAccount
	ringGroups   []RingGroup
	timeConds    []TimeCondition
	recordings   []Recording
	dids         []DID
	haveServers  bool
	haveFwds     bool
	haveVMs      bool
	haveAccounts bool
	haveGroups   bool
	haveTimeCond bool
	haveRecs     bool
	haveDIDs     bool
}

// invalidate drops every cached list. Reads are served from one list per object
// type per run, so a write has to clear them or the read-back after an update
// returns the pre-write values.
func (c *Client) invalidate() {
	c.cache.mu.Lock()
	defer c.cache.mu.Unlock()
	c.cache.servers, c.cache.haveServers = nil, false
	c.cache.forwardings, c.cache.haveFwds = nil, false
	c.cache.voicemails, c.cache.haveVMs = nil, false
	c.cache.subaccounts, c.cache.haveAccounts = nil, false
	c.cache.ringGroups, c.cache.haveGroups = nil, false
	c.cache.timeConds, c.cache.haveTimeCond = nil, false
	c.cache.recordings, c.cache.haveRecs = nil, false
	c.cache.dids, c.cache.haveDIDs = nil, false
}

func (c *Client) cachedServers(load func() ([]Server, error)) ([]Server, error) {
	c.cache.mu.Lock()
	defer c.cache.mu.Unlock()
	if c.cache.haveServers {
		return c.cache.servers, nil
	}
	items, err := load()
	if err != nil {
		return nil, err
	}
	c.cache.servers = items
	c.cache.haveServers = true
	return items, nil
}

func (c *Client) cachedForwardings(load func() ([]Forwarding, error)) ([]Forwarding, error) {
	c.cache.mu.Lock()
	defer c.cache.mu.Unlock()
	if c.cache.haveFwds {
		return c.cache.forwardings, nil
	}
	items, err := load()
	if err != nil {
		return nil, err
	}
	c.cache.forwardings = items
	c.cache.haveFwds = true
	return items, nil
}

func (c *Client) cachedVoicemails(load func() ([]Voicemail, error)) ([]Voicemail, error) {
	c.cache.mu.Lock()
	defer c.cache.mu.Unlock()
	if c.cache.haveVMs {
		return c.cache.voicemails, nil
	}
	items, err := load()
	if err != nil {
		return nil, err
	}
	c.cache.voicemails = items
	c.cache.haveVMs = true
	return items, nil
}

func (c *Client) cachedSubAccounts(load func() ([]SubAccount, error)) ([]SubAccount, error) {
	c.cache.mu.Lock()
	defer c.cache.mu.Unlock()
	if c.cache.haveAccounts {
		return c.cache.subaccounts, nil
	}
	items, err := load()
	if err != nil {
		return nil, err
	}
	c.cache.subaccounts = items
	c.cache.haveAccounts = true
	return items, nil
}

func (c *Client) cachedRingGroups(load func() ([]RingGroup, error)) ([]RingGroup, error) {
	c.cache.mu.Lock()
	defer c.cache.mu.Unlock()
	if c.cache.haveGroups {
		return c.cache.ringGroups, nil
	}
	items, err := load()
	if err != nil {
		return nil, err
	}
	c.cache.ringGroups = items
	c.cache.haveGroups = true
	return items, nil
}

func (c *Client) cachedTimeConditions(load func() ([]TimeCondition, error)) ([]TimeCondition, error) {
	c.cache.mu.Lock()
	defer c.cache.mu.Unlock()
	if c.cache.haveTimeCond {
		return c.cache.timeConds, nil
	}
	items, err := load()
	if err != nil {
		return nil, err
	}
	c.cache.timeConds = items
	c.cache.haveTimeCond = true
	return items, nil
}

func (c *Client) cachedRecordings(load func() ([]Recording, error)) ([]Recording, error) {
	c.cache.mu.Lock()
	defer c.cache.mu.Unlock()
	if c.cache.haveRecs {
		return c.cache.recordings, nil
	}
	items, err := load()
	if err != nil {
		return nil, err
	}
	c.cache.recordings = items
	c.cache.haveRecs = true
	return items, nil
}

func (c *Client) cachedDIDs(load func() ([]DID, error)) ([]DID, error) {
	c.cache.mu.Lock()
	defer c.cache.mu.Unlock()
	if c.cache.haveDIDs {
		return c.cache.dids, nil
	}
	items, err := load()
	if err != nil {
		return nil, err
	}
	c.cache.dids = items
	c.cache.haveDIDs = true
	return items, nil
}
