package client

import (
	"context"
	"fmt"
)

type subAccountsResponse struct {
	Status   string       `json:"status"`
	Accounts []SubAccount `json:"accounts"`
}

// GetSubAccounts lists sub-accounts. account may be an API id, SIP login, or username suffix.
func (c *Client) GetSubAccounts(ctx context.Context, account string) ([]SubAccount, error) {
	items, err := c.cachedSubAccounts(func() ([]SubAccount, error) {
		var resp subAccountsResponse
		if err := c.Call(ctx, "getSubAccounts", map[string]string{}, &resp); err != nil {
			if emptyResult(err) {
				return []SubAccount{}, nil
			}
			return nil, err
		}
		return resp.Accounts, nil
	})
	if err != nil || account == "" {
		return items, err
	}
	return filterList(items, func(x *SubAccount) bool {
		return x.ID.String() == account || x.Account.String() == account || x.Username.String() == account
	}), nil
}

// GetSubAccount returns one sub-account by API id, SIP login (`{main}_{username}`),
// or username suffix. VoIP.ms only accepts the SIP login (or an empty filter) on
// getSubAccounts — numeric ids return status no_subaccount — so we list all and
// match locally.
func (c *Client) GetSubAccount(ctx context.Context, account string) (*SubAccount, error) {
	accounts, err := c.GetSubAccounts(ctx, account)
	if err != nil {
		return nil, err
	}
	if found := matchSubAccount(accounts, account); found != nil {
		return found, nil
	}
	if account == "" {
		return nil, fmt.Errorf("%w: sub-account %s", ErrNotFound, account)
	}
	all, err := c.GetSubAccounts(ctx, "")
	if err != nil {
		return nil, err
	}
	if found := matchSubAccount(all, account); found != nil {
		return found, nil
	}
	return nil, fmt.Errorf("%w: sub-account %s", ErrNotFound, account)
}

func matchSubAccount(accounts []SubAccount, account string) *SubAccount {
	for i := range accounts {
		if accounts[i].ID.String() == account || accounts[i].Account.String() == account || accounts[i].Username.String() == account {
			return &accounts[i]
		}
	}
	if len(accounts) == 1 && account != "" {
		return &accounts[0]
	}
	return nil
}

// CreateSubAccount creates a sub-account. username is the suffix only (not account_suffix).
func (c *Client) CreateSubAccount(ctx context.Context, params map[string]string) error {
	return c.CallWrite(ctx, "createSubAccount", params, nil)
}

// UpdateSubAccount updates a sub-account. params must include id.
func (c *Client) UpdateSubAccount(ctx context.Context, params map[string]string) error {
	return c.CallWrite(ctx, "setSubAccount", params, nil)
}

// DeleteSubAccount deletes a sub-account by numeric API id.
func (c *Client) DeleteSubAccount(ctx context.Context, id string) error {
	return c.CallWrite(ctx, "delSubAccount", map[string]string{"id": id}, nil)
}
