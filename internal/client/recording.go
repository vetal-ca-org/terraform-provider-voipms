package client

import (
	"context"
	"fmt"
)

type recordingsResponse struct {
	Status     string      `json:"status"`
	Recordings []Recording `json:"recordings"`
}

// GetRecordings lists recordings. id, if set, filters to one.
func (c *Client) GetRecordings(ctx context.Context, id string) ([]Recording, error) {
	items, err := c.cachedRecordings(func() ([]Recording, error) {
		var resp recordingsResponse
		if err := c.Call(ctx, "getRecordings", map[string]string{}, &resp); err != nil {
			if emptyResult(err) {
				return []Recording{}, nil
			}
			return nil, err
		}
		return resp.Recordings, nil
	})
	if err != nil || id == "" {
		return items, err
	}
	return filterList(items, func(x *Recording) bool { return x.Recording.String() == id }), nil
}

// GetRecording returns one recording by id.
func (c *Client) GetRecording(ctx context.Context, id string) (*Recording, error) {
	items, err := c.GetRecordings(ctx, id)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].Recording.String() == id {
			return &items[i], nil
		}
	}
	return nil, fmt.Errorf("%w: recording %s", ErrNotFound, id)
}

// FindRecording returns a recording by id or description.
func (c *Client) FindRecording(ctx context.Context, query string) (*Recording, error) {
	items, err := c.GetRecordings(ctx, "")
	if err != nil {
		return nil, err
	}
	return MatchRecording(items, query)
}
