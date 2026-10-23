package client

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type didsResponse struct {
	Status string `json:"status"`
	DIDs   []DID  `json:"dids"`
}

// GetDIDsInfo lists DIDs. did, if set, filters to one number.
func (c *Client) GetDIDsInfo(ctx context.Context, did string) ([]DID, error) {
	items, err := c.cachedDIDs(func() ([]DID, error) {
		var resp didsResponse
		if err := c.Call(ctx, "getDIDsInfo", map[string]string{}, &resp); err != nil {
			if emptyResult(err) {
				return []DID{}, nil
			}
			return nil, err
		}
		return resp.DIDs, nil
	})
	if err != nil || did == "" {
		return items, err
	}
	return filterList(items, func(x *DID) bool { return x.DID.String() == did }), nil
}

// GetDID returns one DID by number.
func (c *Client) GetDID(ctx context.Context, did string) (*DID, error) {
	dids, err := c.GetDIDsInfo(ctx, did)
	if err != nil {
		return nil, err
	}
	for i := range dids {
		if dids[i].DID.String() == did {
			return &dids[i], nil
		}
	}
	if len(dids) == 1 && did != "" {
		return &dids[0], nil
	}
	return nil, fmt.Errorf("%w: DID %s", ErrNotFound, did)
}

// SetDIDInfo updates routing and related DID settings.
func (c *Client) SetDIDInfo(ctx context.Context, params map[string]string) error {
	return c.CallWrite(ctx, "setDIDInfo", params, nil)
}

// SetSMS updates SMS/MMS delivery settings for a DID.
// Calls are serialized and retried on sms_wait_message. That status is
// account-wide: Terraform applying several voipms_did resources at once
// used to wipe webhook_enabled on the losers.
var smsWriteMu sync.Mutex

func (c *Client) SetSMS(ctx context.Context, params map[string]string) error {
	smsWriteMu.Lock()
	defer smsWriteMu.Unlock()

	tries := c.smsWaitTries
	if tries < 1 {
		tries = 1
	}
	var last error
	for i := 0; i < tries; i++ {
		last = c.CallWrite(ctx, "setSMS", params, nil)
		if last == nil {
			return nil
		}
		if !SMSWait(last) {
			return last
		}
		if i == tries-1 {
			break
		}
		d := c.smsWaitDelay
		if d < 0 {
			d = 0
		}
		if err := c.Sleep(ctx, d); err != nil {
			return last
		}
	}
	return last
}

// UseInstantSMSWait is for tests. setSMS retries do not sleep.
func (c *Client) UseInstantSMSWait() {
	c.smsWaitDelay = 0
	c.retryBackoff = func(context.Context, time.Duration) error { return nil }
}

// Sleep waits d, or returns immediately when d is zero. Tests replace the backoff.
func (c *Client) Sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	if c.retryBackoff != nil {
		return c.retryBackoff(ctx, d)
	}
	return sleepContext(ctx, d)
}
