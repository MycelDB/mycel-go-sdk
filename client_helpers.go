package mycel

import (
	"context"

	commonv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/common/v1"
)

func (c *Client) WhoAmI(ctx context.Context) (*PrincipalInfo, error) {
	return DoReadValue(ctx, c, "who am i", func() (*PrincipalInfo, error) {
		callCtx, cancel := c.AuthCallContext(ctx)
		defer cancel()
		res, err := c.Auth.WhoAmI(callCtx, &commonv1.WhoAmIRequest{})
		if err != nil {
			return nil, err
		}
		principal := res.GetPrincipal()
		return &PrincipalInfo{UserID: principal.GetPrincipalId(), PrincipalID: principal.GetPrincipalId(), Username: principal.GetUsername()}, nil
	})
}
