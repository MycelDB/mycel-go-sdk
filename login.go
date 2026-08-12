package mycel

import (
	"context"
	"time"

	commonv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/common/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Client) Login(ctx context.Context, username, password string) (*commonv1.LoginResponse, error) {
	callCtx, cancel := c.CallContext(ctx)
	defer cancel()
	res, err := c.Auth.Login(callCtx, &commonv1.LoginRequest{Username: username, Password: password, Client: c.clientInfo()})
	if err != nil {
		return nil, err
	}
	c.setAuthTokens(res.GetAccessToken(), timestampAsTime(res.GetAccessTokenExpireTime()), res.GetRefreshToken())
	return res, nil
}

func (c *Client) Refresh(ctx context.Context, refreshToken string) (*commonv1.RefreshResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	if refreshToken == "" {
		refreshToken = c.RefreshToken()
	}
	req := &commonv1.RefreshRequest{Client: c.clientInfo()}
	if refreshToken != "" {
		req.RefreshToken = &refreshToken
	}
	res, err := c.Auth.Refresh(callCtx, req)
	if err != nil {
		return nil, err
	}
	c.setAuthTokens(res.GetAccessToken(), timestampAsTime(res.GetAccessTokenExpireTime()), res.GetRefreshToken())
	return res, nil
}

func (c *Client) Logout(ctx context.Context, authSessionID string) (*commonv1.LogoutResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	req := &commonv1.LogoutRequest{}
	if authSessionID != "" {
		req.AuthSessionId = &authSessionID
	}
	res, err := c.Auth.Logout(callCtx, req)
	if err != nil {
		return nil, err
	}
	if authSessionID == "" && c != nil && c.tokens != nil {
		c.tokens.Clear()
	}
	return res, nil
}

func (c *Client) refreshWithStoredToken(ctx context.Context) error {
	_, err := c.Refresh(ctx, "")
	return err
}

func (c *Client) SetAccessToken(token string) {
	if c != nil && c.tokens != nil {
		c.tokens.Set(token)
	}
}

func (c *Client) SetRefreshToken(token string) {
	if c != nil && c.tokens != nil {
		c.tokens.SetRefreshToken(token)
	}
}

func (c *Client) SetAuthTokens(accessToken string, accessTokenExpireTime time.Time, refreshToken string) {
	c.setAuthTokens(accessToken, accessTokenExpireTime, refreshToken)
}

func (c *Client) setAuthTokens(accessToken string, accessTokenExpireTime time.Time, refreshToken string) {
	if c != nil && c.tokens != nil {
		c.tokens.SetTokens(accessToken, accessTokenExpireTime, refreshToken)
	}
}

func (c *Client) AccessToken() string {
	if c == nil || c.tokens == nil {
		return ""
	}
	return c.tokens.Token()
}

func (c *Client) RefreshToken() string {
	if c == nil || c.tokens == nil {
		return ""
	}
	return c.tokens.RefreshToken()
}

func (c *Client) AccessTokenExpireTime() time.Time {
	if c == nil || c.tokens == nil {
		return time.Time{}
	}
	return c.tokens.AccessTokenExpireTime()
}

func (c *Client) AuthContext(ctx context.Context) context.Context {
	if c == nil || c.tokens == nil {
		return ctx
	}
	return c.tokens.Context(ctx)
}

func (c *Client) clientInfo() *commonv1.ClientInfo {
	if c == nil {
		return &commonv1.ClientInfo{Name: "mycel-go-sdk", Platform: "go"}
	}
	return &commonv1.ClientInfo{Name: firstNonEmpty(c.cfg.ClientName, "mycel-go-sdk"), Version: c.cfg.ClientVersion, Platform: firstNonEmpty(c.cfg.Platform, "go"), DeviceLabel: c.cfg.DeviceLabel}
}

func (c *AdminClient) LoginPrincipal(ctx context.Context, username, password string) (*commonv1.LoginResponse, error) {
	callCtx, cancel := c.CallContext(ctx)
	defer cancel()
	res, err := c.Auth.Login(callCtx, &commonv1.LoginRequest{Username: username, Password: password, Client: c.adminClientInfo()})
	if err != nil {
		return nil, err
	}
	c.setAuthTokens(res.GetAccessToken(), timestampAsTime(res.GetAccessTokenExpireTime()), res.GetRefreshToken())
	return res, nil
}

func (c *AdminClient) RefreshPrincipal(ctx context.Context, refreshToken string) (*commonv1.RefreshResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	if refreshToken == "" {
		refreshToken = c.RefreshToken()
	}
	req := &commonv1.RefreshRequest{Client: c.adminClientInfo()}
	if refreshToken != "" {
		req.RefreshToken = &refreshToken
	}
	res, err := c.Auth.Refresh(callCtx, req)
	if err != nil {
		return nil, err
	}
	c.setAuthTokens(res.GetAccessToken(), timestampAsTime(res.GetAccessTokenExpireTime()), res.GetRefreshToken())
	return res, nil
}

func (c *AdminClient) LogoutPrincipal(ctx context.Context, authSessionID string) (*commonv1.LogoutResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	req := &commonv1.LogoutRequest{}
	if authSessionID != "" {
		req.AuthSessionId = &authSessionID
	}
	res, err := c.Auth.Logout(callCtx, req)
	if err != nil {
		return nil, err
	}
	if authSessionID == "" && c != nil && c.tokens != nil {
		c.tokens.Clear()
	}
	return res, nil
}

func (c *AdminClient) refreshWithStoredToken(ctx context.Context) error {
	_, err := c.RefreshPrincipal(ctx, "")
	return err
}

func (c *AdminClient) SetAccessToken(token string) {
	if c != nil && c.tokens != nil {
		c.tokens.Set(token)
	}
}

func (c *AdminClient) SetRefreshToken(token string) {
	if c != nil && c.tokens != nil {
		c.tokens.SetRefreshToken(token)
	}
}

func (c *AdminClient) SetAuthTokens(accessToken string, accessTokenExpireTime time.Time, refreshToken string) {
	c.setAuthTokens(accessToken, accessTokenExpireTime, refreshToken)
}

func (c *AdminClient) setAuthTokens(accessToken string, accessTokenExpireTime time.Time, refreshToken string) {
	if c != nil && c.tokens != nil {
		c.tokens.SetTokens(accessToken, accessTokenExpireTime, refreshToken)
	}
}

func (c *AdminClient) AccessToken() string {
	if c == nil || c.tokens == nil {
		return ""
	}
	return c.tokens.Token()
}

func (c *AdminClient) RefreshToken() string {
	if c == nil || c.tokens == nil {
		return ""
	}
	return c.tokens.RefreshToken()
}

func (c *AdminClient) AccessTokenExpireTime() time.Time {
	if c == nil || c.tokens == nil {
		return time.Time{}
	}
	return c.tokens.AccessTokenExpireTime()
}

func (c *AdminClient) AuthContext(ctx context.Context) context.Context {
	if c == nil || c.tokens == nil {
		return ctx
	}
	return c.tokens.Context(ctx)
}

func (c *AdminClient) adminClientInfo() *commonv1.ClientInfo {
	if c == nil {
		return &commonv1.ClientInfo{Name: "mycel-go-sdk", Platform: "go"}
	}
	return &commonv1.ClientInfo{Name: firstNonEmpty(c.cfg.ClientName, "mycel-go-sdk"), Version: c.cfg.ClientVersion, Platform: firstNonEmpty(c.cfg.Platform, "go"), DeviceLabel: c.cfg.DeviceLabel}
}

func timestampAsTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime().UTC()
}
