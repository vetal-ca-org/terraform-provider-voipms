package client

import (
	"context"
	"fmt"
)

// getTimeConditions answers with the list under a singular key.
type timeConditionsResponse struct {
	Status         string          `json:"status"`
	TimeConditions []TimeCondition `json:"timecondition"`
}

type setTimeConditionResponse struct {
	Status        string     `json:"status"`
	TimeCondition FlexString `json:"timecondition"`
}

// GetTimeConditions lists time conditions. id, if set, filters to one.
func (c *Client) GetTimeConditions(ctx context.Context, id string) ([]TimeCondition, error) {
	items, err := c.cachedTimeConditions(func() ([]TimeCondition, error) {
		var resp timeConditionsResponse
		if err := c.Call(ctx, "getTimeConditions", map[string]string{}, &resp); err != nil {
			if emptyResult(err) {
				return []TimeCondition{}, nil
			}
			return nil, err
		}
		return resp.TimeConditions, nil
	})
	if err != nil || id == "" {
		return items, err
	}
	return filterList(items, func(x *TimeCondition) bool { return x.TimeCondition.String() == id }), nil
}

// GetTimeCondition returns one time condition by id.
func (c *Client) GetTimeCondition(ctx context.Context, id string) (*TimeCondition, error) {
	items, err := c.GetTimeConditions(ctx, id)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].TimeCondition.String() == id {
			return &items[i], nil
		}
	}
	return nil, fmt.Errorf("%w: time condition %s", ErrNotFound, id)
}

// FindTimeCondition returns a time condition by id or name.
func (c *Client) FindTimeCondition(ctx context.Context, query string) (*TimeCondition, error) {
	items, err := c.GetTimeConditions(ctx, "")
	if err != nil {
		return nil, err
	}
	return MatchTimeCondition(items, query)
}

// CreateTimeCondition adds a time condition and returns the id VoIP.ms assigned.
func (c *Client) CreateTimeCondition(ctx context.Context, params map[string]string) (string, error) {
	delete(params, "timecondition")
	var resp setTimeConditionResponse
	if err := c.CallWrite(ctx, "setTimeCondition", params, &resp); err != nil {
		return "", err
	}
	return resp.TimeCondition.String(), nil
}

// UpdateTimeCondition updates an existing one. params must carry timecondition.
func (c *Client) UpdateTimeCondition(ctx context.Context, params map[string]string) error {
	return c.CallWrite(ctx, "setTimeCondition", params, nil)
}

// DeleteTimeCondition deletes a time condition by id.
func (c *Client) DeleteTimeCondition(ctx context.Context, id string) error {
	return c.CallWrite(ctx, "delTimeCondition", map[string]string{"timecondition": id}, nil)
}
