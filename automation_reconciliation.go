package mycel

import (
	"context"

	clientv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/client/v1"
)

// StartGraphAutomationReconciliation starts an idempotent, durable
// reconciliation operation for a graph automation binding. Target discovery
// and invocation scheduling continue asynchronously after this call returns.
func (c *Client) StartGraphAutomationReconciliation(ctx context.Context, spaceID, domainID, bindingID, idempotencyKey string, batchSize, maxConcurrency int32) (*clientv1.GraphAutomationReconciliation, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Automation.StartGraphAutomationReconciliation(callCtx, &clientv1.StartGraphAutomationReconciliationRequest{
		SpaceId: spaceID, DomainId: domainID, BindingId: bindingID, IdempotencyKey: idempotencyKey,
		BatchSize: batchSize, MaxConcurrency: maxConcurrency,
	})
	if err != nil {
		return nil, err
	}
	return res.GetOperation(), nil
}

// GetGraphAutomationReconciliation returns sanitized durable progress for a
// reconciliation operation.
func (c *Client) GetGraphAutomationReconciliation(ctx context.Context, spaceID, operationID string) (*clientv1.GraphAutomationReconciliation, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Automation.GetGraphAutomationReconciliation(callCtx, &clientv1.GetGraphAutomationReconciliationRequest{SpaceId: spaceID, OperationId: operationID})
	if err != nil {
		return nil, err
	}
	return res.GetOperation(), nil
}

// CancelGraphAutomationReconciliation fences new discovery and scheduling for
// a reconciliation operation. Already-running invocations retain normal
// automation execution semantics.
func (c *Client) CancelGraphAutomationReconciliation(ctx context.Context, spaceID, operationID string) (*clientv1.GraphAutomationReconciliation, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Automation.CancelGraphAutomationReconciliation(callCtx, &clientv1.CancelGraphAutomationReconciliationRequest{SpaceId: spaceID, OperationId: operationID})
	if err != nil {
		return nil, err
	}
	return res.GetOperation(), nil
}
