package mycel

import (
	"context"
	"fmt"

	clientv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/client/v1"
)

func (c *Client) ResolveDomainID(ctx context.Context, spaceID, domainKey string) (string, error) {
	if domainKey == "" {
		domainKey = "default"
	}
	return DoReadValue(ctx, c, "resolve domain", func() (string, error) {
		callCtx, cancel := c.AuthCallContext(ctx)
		defer cancel()
		res, err := c.Domain.GetDomain(callCtx, &clientv1.GetDomainRequest{SpaceId: spaceID, Key: domainKey})
		if err != nil {
			return "", fmt.Errorf("resolve domain %q in space %s: %w", domainKey, spaceID, err)
		}
		return res.GetDomain().GetDomainId(), nil
	})
}

func (c *Client) OpenSession(ctx context.Context, spaceID, domainID string) (string, error) {
	return DoReadValue(ctx, c, "open session", func() (string, error) {
		callCtx, cancel := c.AuthCallContext(ctx)
		defer cancel()
		res, err := c.Session.OpenSession(callCtx, &clientv1.OpenSessionRequest{SpaceId: spaceID, DomainId: domainID})
		if err != nil {
			return "", err
		}
		return res.GetSession().GetSessionId(), nil
	})
}

func (c *Client) CloseSession(ctx context.Context, sessionID string) error {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	_, err := c.Session.CloseSession(callCtx, &clientv1.CloseSessionRequest{SessionId: sessionID})
	return err
}

func (c *Client) BeginTransaction(ctx context.Context, sessionID string, mode clientv1.TransactionMode) (string, error) {
	tx, err := c.BeginTransactionWithOperationID(ctx, sessionID, mode, "")
	if err != nil {
		return "", err
	}
	return tx.GetTransactionId(), nil
}

// BeginTransactionWithOperationID begins a transaction with optional client
// operation correlation metadata. Pass NewOperationID() to correlate the write
// with later graph-change events; pass an empty string to let the daemon
// generate an operation ID. The returned transaction includes the resolved
// operation ID.
func (c *Client) BeginTransactionWithOperationID(ctx context.Context, sessionID string, mode clientv1.TransactionMode, operationID string) (*clientv1.GraphTransaction, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Transaction.BeginTransaction(callCtx, &clientv1.BeginTransactionRequest{SessionId: sessionID, Mode: mode, OperationId: operationID})
	if err != nil {
		return nil, err
	}
	return res.GetTransaction(), nil
}

func (c *Client) BeginReadWriteTransaction(ctx context.Context, sessionID string) (string, error) {
	return c.BeginTransaction(ctx, sessionID, clientv1.TransactionMode_TRANSACTION_MODE_READ_WRITE)
}

func (c *Client) BeginReadWriteTransactionWithOperationID(ctx context.Context, sessionID string, operationID string) (*clientv1.GraphTransaction, error) {
	return c.BeginTransactionWithOperationID(ctx, sessionID, clientv1.TransactionMode_TRANSACTION_MODE_READ_WRITE, operationID)
}

func (c *Client) BeginReadOnlyTransaction(ctx context.Context, sessionID string) (string, error) {
	return c.BeginTransaction(ctx, sessionID, clientv1.TransactionMode_TRANSACTION_MODE_READ_ONLY)
}

func (c *Client) BeginReadOnlyTransactionWithOperationID(ctx context.Context, sessionID string, operationID string) (*clientv1.GraphTransaction, error) {
	return c.BeginTransactionWithOperationID(ctx, sessionID, clientv1.TransactionMode_TRANSACTION_MODE_READ_ONLY, operationID)
}

func (c *Client) CommitTransaction(ctx context.Context, txID string) error {
	_, err := c.CommitTransactionResult(ctx, txID)
	return err
}

// CommitTransactionResult commits a read-write transaction and returns commit
// metadata, including the operation ID associated with the transaction.
func (c *Client) CommitTransactionResult(ctx context.Context, txID string) (*clientv1.TransactionCommit, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Transaction.CommitTransaction(callCtx, &clientv1.CommitTransactionRequest{TransactionId: txID})
	if err != nil {
		return nil, err
	}
	return res.GetCommit(), nil
}

func (c *Client) CloseTransaction(ctx context.Context, txID string) error {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	_, err := c.Transaction.CloseTransaction(callCtx, &clientv1.CloseTransactionRequest{TransactionId: txID})
	return err
}
