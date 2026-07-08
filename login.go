package mycel

import (
	"context"

	adminv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/admin/v1"
	clientv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/client/v1"
)

func (c *Client) Login(ctx context.Context, username, password string) (*clientv1.LoginResponse, error) {
	callCtx, cancel := c.CallContext(ctx)
	defer cancel()
	res, err := c.Auth.Login(callCtx, &clientv1.LoginRequest{Username: username, Password: password, Client: c.clientInfo()})
	if err != nil {
		return nil, err
	}
	c.SetAccessToken(res.GetAccessToken())
	return res, nil
}

func (c *Client) Refresh(ctx context.Context, refreshToken string) (*clientv1.RefreshResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	req := &clientv1.RefreshRequest{Client: c.clientInfo()}
	if refreshToken != "" {
		req.RefreshToken = &refreshToken
	}
	res, err := c.Auth.Refresh(callCtx, req)
	if err != nil {
		return nil, err
	}
	c.SetAccessToken(res.GetAccessToken())
	return res, nil
}

func (c *Client) SetAccessToken(token string) {
	if c != nil && c.tokens != nil {
		c.tokens.Set(token)
	}
}

func (c *Client) AccessToken() string {
	if c == nil || c.tokens == nil {
		return ""
	}
	return c.tokens.Token()
}

func (c *Client) AuthContext(ctx context.Context) context.Context {
	if c == nil || c.tokens == nil {
		return ctx
	}
	return c.tokens.Context(ctx)
}

func (c *Client) clientInfo() *clientv1.ClientInfo {
	if c == nil {
		return &clientv1.ClientInfo{Name: "mycel-go-sdk", Platform: "go"}
	}
	return &clientv1.ClientInfo{Name: firstNonEmpty(c.cfg.ClientName, "mycel-go-sdk"), Version: c.cfg.ClientVersion, Platform: firstNonEmpty(c.cfg.Platform, "go"), DeviceLabel: c.cfg.DeviceLabel}
}

func (c *AdminClient) LoginOperator(ctx context.Context, username, password string) (*adminv1.LoginOperatorResponse, error) {
	callCtx, cancel := c.CallContext(ctx)
	defer cancel()
	res, err := c.Auth.LoginOperator(callCtx, &adminv1.LoginOperatorRequest{Username: username, Password: password, Client: c.operatorClientInfo()})
	if err != nil {
		return nil, err
	}
	c.SetAccessToken(res.GetAccessToken())
	return res, nil
}

func (c *AdminClient) SetAccessToken(token string) {
	if c != nil && c.tokens != nil {
		c.tokens.Set(token)
	}
}

func (c *AdminClient) AccessToken() string {
	if c == nil || c.tokens == nil {
		return ""
	}
	return c.tokens.Token()
}

func (c *AdminClient) AuthContext(ctx context.Context) context.Context {
	if c == nil || c.tokens == nil {
		return ctx
	}
	return c.tokens.Context(ctx)
}

func (c *AdminClient) operatorClientInfo() *adminv1.OperatorClientInfo {
	if c == nil {
		return &adminv1.OperatorClientInfo{Name: "mycel-go-sdk", Platform: "go"}
	}
	return &adminv1.OperatorClientInfo{Name: firstNonEmpty(c.cfg.ClientName, "mycel-go-sdk"), Version: c.cfg.ClientVersion, Platform: firstNonEmpty(c.cfg.Platform, "go"), DeviceLabel: c.cfg.DeviceLabel}
}
