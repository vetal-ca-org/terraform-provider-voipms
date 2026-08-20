package client

import (
	"context"
	"fmt"
)

type callerIDFilteringResponse struct {
	Status    string           `json:"status"`
	Filtering []CallerIDFilter `json:"filtering"`
}

// GetCallerIDFilters lists caller-ID filter rules. id, if set, filters to one rule.
func (c *Client) GetCallerIDFilters(ctx context.Context, id string) ([]CallerIDFilter, error) {
	params := map[string]string{}
	if id != "" {
		params["filtering"] = id
	}
	var resp callerIDFilteringResponse
	err := c.Call(ctx, "getCallerIDFiltering", params, &resp)
	if err != nil {
		if emptyResult(err) {
			return []CallerIDFilter{}, nil
		}
		return nil, err
	}
	return resp.Filtering, nil
}

// GetCallerIDFilter returns one filter by id.
func (c *Client) GetCallerIDFilter(ctx context.Context, id string) (*CallerIDFilter, error) {
	items, err := c.GetCallerIDFilters(ctx, id)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].Filtering.String() == id {
			return &items[i], nil
		}
	}
	if len(items) == 1 && id != "" {
		return &items[0], nil
	}
	return nil, fmt.Errorf("%w: caller ID filter %s", ErrNotFound, id)
}

// SetCallerIDFilter creates (no filtering id) or updates a filter.
func (c *Client) SetCallerIDFilter(ctx context.Context, params map[string]string) error {
	return c.CallWrite(ctx, "setCallerIDFiltering", params, nil)
}

// DeleteCallerIDFilter deletes a filter by id.
func (c *Client) DeleteCallerIDFilter(ctx context.Context, id string) error {
	return c.Call(ctx, "delCallerIDFiltering", map[string]string{"filtering": id}, nil)
}
