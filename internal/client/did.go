package client

import (
	"context"
	"fmt"
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
func (c *Client) SetSMS(ctx context.Context, params map[string]string) error {
	return c.CallWrite(ctx, "setSMS", params, nil)
}
