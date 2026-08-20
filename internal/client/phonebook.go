package client

import (
	"context"
	"fmt"
)

type phonebookResponse struct {
	Status     string           `json:"status"`
	Phonebooks []PhonebookEntry `json:"phonebooks"`
}

type phonebookGroupsResponse struct {
	Status          string           `json:"status"`
	PhonebookGroups []PhonebookGroup `json:"phonebook_groups"`
}

// GetPhonebook lists phonebook entries. id, if set, filters to one entry.
func (c *Client) GetPhonebook(ctx context.Context, id string) ([]PhonebookEntry, error) {
	params := map[string]string{}
	if id != "" {
		params["phonebook"] = id
	}
	var resp phonebookResponse
	err := c.Call(ctx, "getPhonebook", params, &resp)
	if err != nil {
		if emptyResult(err) {
			return []PhonebookEntry{}, nil
		}
		return nil, err
	}
	return resp.Phonebooks, nil
}

// GetPhonebookEntry returns one phonebook entry by id.
func (c *Client) GetPhonebookEntry(ctx context.Context, id string) (*PhonebookEntry, error) {
	items, err := c.GetPhonebook(ctx, id)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].Phonebook.String() == id {
			return &items[i], nil
		}
	}
	if len(items) == 1 && id != "" {
		return &items[0], nil
	}
	return nil, fmt.Errorf("%w: phonebook entry %s", ErrNotFound, id)
}

// SetPhonebook creates (no phonebook id) or updates an entry.
func (c *Client) SetPhonebook(ctx context.Context, params map[string]string) error {
	return c.CallWrite(ctx, "setPhonebook", params, nil)
}

// DeletePhonebook deletes a phonebook entry by id.
func (c *Client) DeletePhonebook(ctx context.Context, id string) error {
	return c.Call(ctx, "delPhonebook", map[string]string{"phonebook": id}, nil)
}

// GetPhonebookGroups lists phonebook groups. id, if set, filters to one group.
func (c *Client) GetPhonebookGroups(ctx context.Context, id string) ([]PhonebookGroup, error) {
	params := map[string]string{}
	if id != "" {
		params["group"] = id
	}
	var resp phonebookGroupsResponse
	err := c.Call(ctx, "getPhonebookGroups", params, &resp)
	if err != nil {
		if emptyResult(err) {
			return []PhonebookGroup{}, nil
		}
		return nil, err
	}
	return resp.PhonebookGroups, nil
}

// GetPhonebookGroup returns one group by id.
func (c *Client) GetPhonebookGroup(ctx context.Context, id string) (*PhonebookGroup, error) {
	items, err := c.GetPhonebookGroups(ctx, id)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].PhonebookGroup.String() == id {
			return &items[i], nil
		}
	}
	if len(items) == 1 && id != "" {
		return &items[0], nil
	}
	return nil, fmt.Errorf("%w: phonebook group %s", ErrNotFound, id)
}

// SetPhonebookGroup creates (no group id) or updates a group.
func (c *Client) SetPhonebookGroup(ctx context.Context, params map[string]string) error {
	return c.CallWrite(ctx, "setPhonebookGroup", params, nil)
}

// DeletePhonebookGroup deletes a group by id.
func (c *Client) DeletePhonebookGroup(ctx context.Context, id string) error {
	return c.Call(ctx, "delPhonebookGroup", map[string]string{"group": id}, nil)
}
