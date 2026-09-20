package client

import (
	"context"
	"fmt"
)

type ringGroupsResponse struct {
	Status     string      `json:"status"`
	RingGroups []RingGroup `json:"ring_groups"`
}

type setRingGroupResponse struct {
	Status    string     `json:"status"`
	RingGroup FlexString `json:"ring_group"`
}

// GetRingGroups lists ring groups. id, if set, filters to one group.
func (c *Client) GetRingGroups(ctx context.Context, id string) ([]RingGroup, error) {
	items, err := c.cachedRingGroups(func() ([]RingGroup, error) {
		var resp ringGroupsResponse
		if err := c.Call(ctx, "getRingGroups", map[string]string{}, &resp); err != nil {
			if emptyResult(err) {
				return []RingGroup{}, nil
			}
			return nil, err
		}
		return resp.RingGroups, nil
	})
	if err != nil || id == "" {
		return items, err
	}
	return filterList(items, func(x *RingGroup) bool { return x.RingGroup.String() == id }), nil
}

// GetRingGroup returns one ring group by id.
func (c *Client) GetRingGroup(ctx context.Context, id string) (*RingGroup, error) {
	items, err := c.GetRingGroups(ctx, id)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].RingGroup.String() == id {
			return &items[i], nil
		}
	}
	if len(items) == 1 && id != "" {
		return &items[0], nil
	}
	return nil, fmt.Errorf("%w: ring group %s", ErrNotFound, id)
}

// FindRingGroup returns a ring group by id or name.
func (c *Client) FindRingGroup(ctx context.Context, query string) (*RingGroup, error) {
	items, err := c.GetRingGroups(ctx, "")
	if err != nil {
		return nil, err
	}
	return MatchRingGroup(items, query)
}

// CreateRingGroup adds a ring group and returns the id VoIP.ms assigned.
func (c *Client) CreateRingGroup(ctx context.Context, params map[string]string) (string, error) {
	delete(params, "ring_group")
	var resp setRingGroupResponse
	if err := c.CallWrite(ctx, "setRingGroup", params, &resp); err != nil {
		return "", err
	}
	return resp.RingGroup.String(), nil
}

// UpdateRingGroup updates an existing ring group. params must carry ring_group.
func (c *Client) UpdateRingGroup(ctx context.Context, params map[string]string) error {
	return c.CallWrite(ctx, "setRingGroup", params, nil)
}

// DeleteRingGroup deletes a ring group by id. delRingGroup spells the
// parameter `ringgroup`, without the underscore the other calls use.
func (c *Client) DeleteRingGroup(ctx context.Context, id string) error {
	return c.CallWrite(ctx, "delRingGroup", map[string]string{"ringgroup": id}, nil)
}
