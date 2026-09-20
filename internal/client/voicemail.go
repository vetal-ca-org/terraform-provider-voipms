package client

import (
	"context"
	"fmt"
)

type voicemailsResponse struct {
	Status     string      `json:"status"`
	Voicemails []Voicemail `json:"voicemails"`
}

// GetVoicemails lists voicemail boxes. mailbox, if set, filters to one box.
func (c *Client) GetVoicemails(ctx context.Context, mailbox string) ([]Voicemail, error) {
	items, err := c.cachedVoicemails(func() ([]Voicemail, error) {
		var resp voicemailsResponse
		if err := c.Call(ctx, "getVoicemails", map[string]string{}, &resp); err != nil {
			if emptyResult(err) {
				return []Voicemail{}, nil
			}
			return nil, err
		}
		return resp.Voicemails, nil
	})
	if err != nil || mailbox == "" {
		return items, err
	}
	return filterList(items, func(x *Voicemail) bool { return x.Mailbox.String() == mailbox }), nil
}

// GetVoicemail returns one mailbox.
func (c *Client) GetVoicemail(ctx context.Context, mailbox string) (*Voicemail, error) {
	items, err := c.GetVoicemails(ctx, mailbox)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].Mailbox.String() == mailbox {
			return &items[i], nil
		}
	}
	if len(items) == 1 && mailbox != "" {
		return &items[0], nil
	}
	return nil, fmt.Errorf("%w: voicemail %s", ErrNotFound, mailbox)
}

// FindVoicemail returns a mailbox by number or display name.
func (c *Client) FindVoicemail(ctx context.Context, query string) (*Voicemail, error) {
	items, err := c.GetVoicemails(ctx, "")
	if err != nil {
		return nil, err
	}
	return MatchVoicemail(items, query)
}

// CreateVoicemail creates a mailbox. params must include digits (mailbox number).
func (c *Client) CreateVoicemail(ctx context.Context, params map[string]string) error {
	return c.CallWrite(ctx, "createVoicemail", params, nil)
}

// UpdateVoicemail updates a mailbox.
func (c *Client) UpdateVoicemail(ctx context.Context, params map[string]string) error {
	return c.CallWrite(ctx, "setVoicemail", params, nil)
}

// DeleteVoicemail deletes a mailbox.
func (c *Client) DeleteVoicemail(ctx context.Context, mailbox string) error {
	return c.CallWrite(ctx, "delVoicemail", map[string]string{"mailbox": mailbox}, nil)
}
