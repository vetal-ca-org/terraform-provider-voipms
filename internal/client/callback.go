package client

import (
	"context"
	"fmt"
)

type callbacksResponse struct {
	Status    string     `json:"status"`
	Callbacks []Callback `json:"callbacks"`
}

// GetCallbacks lists callbacks. id, if set, filters to one callback code.
func (c *Client) GetCallbacks(ctx context.Context, id string) ([]Callback, error) {
	params := map[string]string{}
	if id != "" {
		params["callback"] = id
	}
	var resp callbacksResponse
	err := c.Call(ctx, "getCallbacks", params, &resp)
	if err != nil {
		if emptyResult(err) {
			return []Callback{}, nil
		}
		return nil, err
	}
	return resp.Callbacks, nil
}

// GetCallback returns one callback by id.
func (c *Client) GetCallback(ctx context.Context, id string) (*Callback, error) {
	items, err := c.GetCallbacks(ctx, id)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].Callback.String() == id {
			return &items[i], nil
		}
	}
	if len(items) == 1 && id != "" {
		return &items[0], nil
	}
	return nil, fmt.Errorf("%w: callback %s", ErrNotFound, id)
}

// FindCallback returns a callback by id or description.
func (c *Client) FindCallback(ctx context.Context, query string) (*Callback, error) {
	items, err := c.GetCallbacks(ctx, "")
	if err != nil {
		return nil, err
	}
	return MatchCallback(items, query)
}

// SetCallback creates (no callback id) or updates a callback.
func (c *Client) SetCallback(ctx context.Context, params map[string]string) error {
	return c.CallWrite(ctx, "setCallback", params, nil)
}

// DeleteCallback deletes a callback by id.
func (c *Client) DeleteCallback(ctx context.Context, id string) error {
	return c.CallWrite(ctx, "delCallback", map[string]string{"callback": id}, nil)
}
