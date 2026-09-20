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
	items, err := c.cachedForwardings(func() ([]Forwarding, error) {
		var resp forwardingsResponse
		if err := c.Call(ctx, "getForwardings", map[string]string{}, &resp); err != nil {
			if emptyResult(err) {
				return []Forwarding{}, nil
			}
			return nil, err
		}
		return resp.Forwardings, nil
	})
	if err != nil || id == "" {
		return items, err
	}
	return filterList(items, func(x *Forwarding) bool { return x.Forwarding.String() == id }), nil
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

// FindForwarding returns a forwarding by id or description.
func (c *Client) FindForwarding(ctx context.Context, query string) (*Forwarding, error) {
	items, err := c.GetForwardings(ctx, "")
	if err != nil {
		return nil, err
	}
	return MatchForwarding(items, query)
}

// SetForwarding creates (no forwarding id) or updates a forwarding.
func (c *Client) SetForwarding(ctx context.Context, params map[string]string) error {
	return c.CallWrite(ctx, "setForwarding", params, nil)
}

// DeleteForwarding deletes a forwarding by id.
func (c *Client) DeleteForwarding(ctx context.Context, id string) error {
	return c.CallWrite(ctx, "delForwarding", map[string]string{"forwarding": id}, nil)
}
