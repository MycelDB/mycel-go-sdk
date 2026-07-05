package mycel

import (
	"context"
	"fmt"

	clientv1 "github.com/myceldb/mycel-api/gen/go/mycel/client/v1"
)

func (c *Client) ResolveDomainID(ctx context.Context, spaceID, domainKey string) (string, error) {
	if domainKey == "" {
		domainKey = "default"
	}
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Domain.GetDomain(callCtx, &clientv1.GetDomainRequest{SpaceId: spaceID, Key: domainKey})
	if err != nil {
		return "", fmt.Errorf("resolve domain %q in space %s: %w", domainKey, spaceID, err)
	}
	return res.GetDomain().GetDomainId(), nil
}

func (c *Client) OpenSession(ctx context.Context, spaceID, domainID string) (string, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Session.OpenSession(callCtx, &clientv1.OpenSessionRequest{SpaceId: spaceID, DomainId: domainID})
	if err != nil {
		return "", err
	}
	return res.GetSession().GetSessionId(), nil
}

func (c *Client) CloseSession(ctx context.Context, sessionID string) error {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	_, err := c.Session.CloseSession(callCtx, &clientv1.CloseSessionRequest{SessionId: sessionID})
	return err
}

func (c *Client) BeginTransaction(ctx context.Context, sessionID string, mode clientv1.TransactionMode) (string, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Transaction.BeginTransaction(callCtx, &clientv1.BeginTransactionRequest{SessionId: sessionID, Mode: mode})
	if err != nil {
		return "", err
	}
	return res.GetTransaction().GetTransactionId(), nil
}

func (c *Client) BeginReadWriteTransaction(ctx context.Context, sessionID string) (string, error) {
	return c.BeginTransaction(ctx, sessionID, clientv1.TransactionMode_TRANSACTION_MODE_READ_WRITE)
}

func (c *Client) BeginReadOnlyTransaction(ctx context.Context, sessionID string) (string, error) {
	return c.BeginTransaction(ctx, sessionID, clientv1.TransactionMode_TRANSACTION_MODE_READ_ONLY)
}

func (c *Client) CommitTransaction(ctx context.Context, txID string) error {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	_, err := c.Transaction.CommitTransaction(callCtx, &clientv1.CommitTransactionRequest{TransactionId: txID})
	return err
}

func (c *Client) CloseTransaction(ctx context.Context, txID string) error {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	_, err := c.Transaction.CloseTransaction(callCtx, &clientv1.CloseTransactionRequest{TransactionId: txID})
	return err
}
