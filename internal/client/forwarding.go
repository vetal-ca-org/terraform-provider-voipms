package client

import (
	"context"
	"fmt"
)

type forwardingsResponse struct {
	Status      string       `json:"status"`
	Forwardings []Forwarding `json:"forwardings"`
}

// GetForwardings lists call forwardings. id, if set, filters to one forwarding code.
func (c *Client) GetForwardings(ctx context.Context, id string) ([]Forwarding, error) {
	params := map[string]string{}
	if id != "" {
		params["forwarding"] = id
	}
	var resp forwardingsResponse
	err := c.Call(ctx, "getForwardings", params, &resp)
	if err != nil {
		if emptyResult(err) {
			return []Forwarding{}, nil
		}
		return nil, err
	}
	return resp.Forwardings, nil
}

// GetForwarding returns one forwarding by id.
func (c *Client) GetForwarding(ctx context.Context, id string) (*Forwarding, error) {
	items, err := c.GetForwardings(ctx, id)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].Forwarding.String() == id {
			return &items[i], nil
		}
	}
	if len(items) == 1 && id != "" {
		return &items[0], nil
	}
	return nil, fmt.Errorf("%w: forwarding %s", ErrNotFound, id)
}

// SetForwarding creates (no forwarding id) or updates a forwarding.
func (c *Client) SetForwarding(ctx context.Context, params map[string]string) error {
	return c.CallWrite(ctx, "setForwarding", params, nil)
}

// DeleteForwarding deletes a forwarding by id.
func (c *Client) DeleteForwarding(ctx context.Context, id string) error {
	return c.Call(ctx, "delForwarding", map[string]string{"forwarding": id}, nil)
}
