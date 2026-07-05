package mycel

import (
	"context"
)

func (c *Client) CallContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if c != nil && c.cfg.CallTimeout > 0 {
		return context.WithTimeout(ctx, c.cfg.CallTimeout)
	}
	return context.WithCancel(ctx)
}

func (c *Client) AuthCallContext(ctx context.Context) (context.Context, context.CancelFunc) {
	callCtx, cancel := c.CallContext(ctx)
	return c.AuthContext(callCtx), cancel
}

func (c *AdminClient) CallContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if c != nil && c.cfg.CallTimeout > 0 {
		return context.WithTimeout(ctx, c.cfg.CallTimeout)
	}
	return context.WithCancel(ctx)
}

func (c *AdminClient) AuthCallContext(ctx context.Context) (context.Context, context.CancelFunc) {
	callCtx, cancel := c.CallContext(ctx)
	return c.AuthContext(callCtx), cancel
}
