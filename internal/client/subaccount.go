package client

import (
	"context"
	"fmt"
)

type subAccountsResponse struct {
	Status   string       `json:"status"`
	Accounts []SubAccount `json:"accounts"`
}

// GetSubAccounts lists sub-accounts. account may be an API id or full SIP login.
func (c *Client) GetSubAccounts(ctx context.Context, account string) ([]SubAccount, error) {
	params := map[string]string{}
	if account != "" {
		params["account"] = account
	}
	var resp subAccountsResponse
	err := c.Call(ctx, "getSubAccounts", params, &resp)
	if err != nil {
		if emptyResult(err) {
			return []SubAccount{}, nil
		}
		return nil, err
	}
	return resp.Accounts, nil
}

// GetSubAccount returns one sub-account by API id or SIP login (`{main}_{username}`).
func (c *Client) GetSubAccount(ctx context.Context, account string) (*SubAccount, error) {
	accounts, err := c.GetSubAccounts(ctx, account)
	if err != nil {
		return nil, err
	}
	for i := range accounts {
		if accounts[i].ID.String() == account || accounts[i].Account.String() == account || accounts[i].Username.String() == account {
			return &accounts[i], nil
		}
	}
	if len(accounts) == 1 && account != "" {
		return &accounts[0], nil
	}
	return nil, fmt.Errorf("%w: sub-account %s", ErrNotFound, account)
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
	return c.Call(ctx, "delSubAccount", map[string]string{"id": id}, nil)
}
