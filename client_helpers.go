package mycel

import (
	"context"

	clientv1 "github.com/myceldb/mycel-api/gen/go/mycel/client/v1"
)

func (c *Client) WhoAmI(ctx context.Context) (*PrincipalInfo, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Auth.WhoAmI(callCtx, &clientv1.WhoAmIRequest{})
	if err != nil {
		return nil, err
	}
	principal := res.GetPrincipal()
	return &PrincipalInfo{UserID: principal.GetUserId(), Username: principal.GetUsername()}, nil
}
