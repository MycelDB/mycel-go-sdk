package mycel

import (
	"context"

	adminv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/admin/v1"
)

func (c *AdminClient) GetUsage(ctx context.Context, req *adminv1.GetUsageRequest) (*adminv1.GetUsageResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	if req == nil {
		req = &adminv1.GetUsageRequest{}
	}
	return c.Usage.GetUsage(callCtx, req)
}

func (c *AdminClient) BatchGetUsage(ctx context.Context, req *adminv1.BatchGetUsageRequest) (*adminv1.BatchGetUsageResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	if req == nil {
		req = &adminv1.BatchGetUsageRequest{}
	}
	return c.Usage.BatchGetUsage(callCtx, req)
}

func (c *AdminClient) ListUsage(ctx context.Context, req *adminv1.ListUsageRequest) (*adminv1.ListUsageResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	if req == nil {
		req = &adminv1.ListUsageRequest{}
	}
	return c.Usage.ListUsage(callCtx, req)
}

func (c *AdminClient) GetUsageAccountingHealth(ctx context.Context, req *adminv1.GetUsageAccountingHealthRequest) (*adminv1.GetUsageAccountingHealthResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	if req == nil {
		req = &adminv1.GetUsageAccountingHealthRequest{}
	}
	return c.Usage.GetUsageAccountingHealth(callCtx, req)
}

func (c *AdminClient) UpsertUsageLimit(ctx context.Context, req *adminv1.UpsertUsageLimitRequest) (*adminv1.UpsertUsageLimitResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	if req == nil {
		req = &adminv1.UpsertUsageLimitRequest{}
	}
	return c.UsageAdmission.UpsertUsageLimit(callCtx, req)
}

func (c *AdminClient) DeleteUsageLimit(ctx context.Context, req *adminv1.DeleteUsageLimitRequest) (*adminv1.DeleteUsageLimitResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	if req == nil {
		req = &adminv1.DeleteUsageLimitRequest{}
	}
	return c.UsageAdmission.DeleteUsageLimit(callCtx, req)
}

func (c *AdminClient) GetUsageLimit(ctx context.Context, req *adminv1.GetUsageLimitRequest) (*adminv1.GetUsageLimitResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	if req == nil {
		req = &adminv1.GetUsageLimitRequest{}
	}
	return c.UsageAdmission.GetUsageLimit(callCtx, req)
}

func (c *AdminClient) ListUsageLimits(ctx context.Context, req *adminv1.ListUsageLimitsRequest) (*adminv1.ListUsageLimitsResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	if req == nil {
		req = &adminv1.ListUsageLimitsRequest{}
	}
	return c.UsageAdmission.ListUsageLimits(callCtx, req)
}
