package mycel

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

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

// IdempotentTransaction carries the durable commit scope chosen before a
// read-write transaction begins. The key is bound only when Commit succeeds.
type IdempotentTransaction struct {
	Transaction   *clientv1.GraphTransaction
	CommitOptions CommitTransactionOptions
}

// BeginIdempotentReadWriteTransaction begins a read-write transaction and
// retains the commit scope that must be reused after an ambiguous outcome.
func (c *Client) BeginIdempotentReadWriteTransaction(ctx context.Context, sessionID, spaceID, domainID, key string) (*IdempotentTransaction, error) {
	return c.BeginIdempotentReadWriteTransactionWithOperationID(ctx, sessionID, spaceID, domainID, key, "")
}

func (c *Client) BeginIdempotentReadWriteTransactionWithOperationID(ctx context.Context, sessionID, spaceID, domainID, key, operationID string) (*IdempotentTransaction, error) {
	options := CommitTransactionOptions{IdempotencyKey: key, SpaceID: spaceID, DomainID: domainID}
	if err := validateCommitTransactionOptions(options); err != nil {
		return nil, err
	}
	tx, err := c.BeginReadWriteTransactionWithOperationID(ctx, sessionID, operationID)
	if err != nil {
		return nil, err
	}
	return &IdempotentTransaction{Transaction: tx, CommitOptions: options}, nil
}

// CommitIdempotentTransaction commits a transaction returned by
// BeginIdempotentReadWriteTransaction.
func (c *Client) CommitIdempotentTransaction(ctx context.Context, tx *IdempotentTransaction) (*clientv1.CommitTransactionResponse, error) {
	if tx == nil || tx.Transaction == nil || tx.Transaction.GetTransactionId() == "" {
		return nil, fmt.Errorf("idempotent transaction is required")
	}
	return c.CommitTransactionWithOptions(ctx, tx.Transaction.GetTransactionId(), tx.CommitOptions)
}

func (c *Client) CommitTransaction(ctx context.Context, txID string) error {
	_, err := c.CommitTransactionResult(ctx, txID)
	return err
}

// CommitTransactionOptions controls durable transaction-level idempotency.
// IdempotencyKey is scoped to the authenticated principal, SpaceID, and
// DomainID. Keep the same key and scope when retrying an ambiguous commit.
type CommitTransactionOptions struct {
	IdempotencyKey string
	SpaceID        string
	DomainID       string
}

// CommitTransactionResult commits a read-write transaction and returns commit
// metadata, including the operation ID associated with the transaction.
func (c *Client) CommitTransactionResult(ctx context.Context, txID string) (*clientv1.TransactionCommit, error) {
	res, err := c.CommitTransactionWithOptions(ctx, txID, CommitTransactionOptions{})
	if err != nil {
		return nil, err
	}
	return res.GetCommit(), nil
}

// CommitTransactionIdempotent durably binds the transaction's canonical graph
// mutation to key within the authenticated principal, space, and domain scope.
// A retry with the same mutation returns the original commit metadata and sets
// idempotent_replay; reuse with a different mutation fails.
func (c *Client) CommitTransactionIdempotent(ctx context.Context, txID, spaceID, domainID, key string) (*clientv1.CommitTransactionResponse, error) {
	return c.CommitTransactionWithOptions(ctx, txID, CommitTransactionOptions{IdempotencyKey: key, SpaceID: spaceID, DomainID: domainID})
}

// CommitTransactionWithOptions commits a transaction using the supplied
// idempotency options. It does not invent or rotate keys, so callers can safely
// retry an ambiguous call with the exact same options.
func (c *Client) CommitTransactionWithOptions(ctx context.Context, txID string, options CommitTransactionOptions) (*clientv1.CommitTransactionResponse, error) {
	if err := validateCommitTransactionOptions(options); err != nil {
		return nil, err
	}
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Transaction.CommitTransaction(callCtx, &clientv1.CommitTransactionRequest{
		TransactionId:  txID,
		IdempotencyKey: options.IdempotencyKey,
		SpaceId:        options.SpaceID,
		DomainId:       options.DomainID,
	})
}

func validateCommitTransactionOptions(options CommitTransactionOptions) error {
	if options.IdempotencyKey == "" {
		if options.SpaceID != "" || options.DomainID != "" {
			return fmt.Errorf("space ID and domain ID require an idempotency key")
		}
		return nil
	}
	if options.IdempotencyKey != strings.TrimSpace(options.IdempotencyKey) || !utf8.ValidString(options.IdempotencyKey) || len(options.IdempotencyKey) > 128 {
		return fmt.Errorf("idempotency key must be valid UTF-8 without surrounding whitespace and at most 128 bytes")
	}
	if strings.TrimSpace(options.SpaceID) == "" || strings.TrimSpace(options.DomainID) == "" {
		return fmt.Errorf("space ID and domain ID are required with an idempotency key")
	}
	return nil
}

func (c *Client) CloseTransaction(ctx context.Context, txID string) error {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	_, err := c.Transaction.CloseTransaction(callCtx, &clientv1.CloseTransactionRequest{TransactionId: txID})
	return err
}
