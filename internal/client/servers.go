package client

import (
	"context"
	"fmt"
)

type serversResponse struct {
	Status  string   `json:"status"`
	Servers []Server `json:"servers"`
}

// GetServersInfo lists VoIP.ms POPs. pop, if set, filters to one server_pop.
func (c *Client) GetServersInfo(ctx context.Context, pop string) ([]Server, error) {
	params := map[string]string{}
	if pop != "" {
		params["server_pop"] = pop
	}
	var resp serversResponse
	err := c.Call(ctx, "getServersInfo", params, &resp)
	if err != nil {
		if emptyResult(err) {
			return []Server{}, nil
		}
		return nil, err
	}
	return resp.Servers, nil
}

// GetServer returns one POP by server_pop id.
func (c *Client) GetServer(ctx context.Context, pop string) (*Server, error) {
	items, err := c.GetServersInfo(ctx, pop)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].POP.String() == pop {
			return &items[i], nil
		}
	}
	if len(items) == 1 && pop != "" {
		return &items[0], nil
	}
	return nil, fmt.Errorf("%w: server pop %s", ErrNotFound, pop)
}
