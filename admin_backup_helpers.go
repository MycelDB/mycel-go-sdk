package mycel

import (
	"context"

	adminv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/admin/v1"
)

func (c *AdminClient) GetBackupPolicy(ctx context.Context) (*adminv1.BackupPolicy, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Backup.GetBackupPolicy(callCtx, &adminv1.GetBackupPolicyRequest{})
	if err != nil {
		return nil, err
	}
	return res.GetPolicy(), nil
}

func (c *AdminClient) UpdateBackupPolicy(ctx context.Context, policy *adminv1.BackupPolicy) (*adminv1.BackupPolicy, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Backup.UpdateBackupPolicy(callCtx, &adminv1.UpdateBackupPolicyRequest{Policy: policy})
	if err != nil {
		return nil, err
	}
	return res.GetPolicy(), nil
}

func (c *AdminClient) ListBackupRepositories(ctx context.Context, pageSize int32, pageToken string) (*adminv1.ListBackupRepositoriesResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.ListBackupRepositories(callCtx, &adminv1.ListBackupRepositoriesRequest{PageSize: pageSize, PageToken: pageToken})
}

func (c *AdminClient) GetBackupRepository(ctx context.Context, repositoryID string) (*adminv1.BackupRepository, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Backup.GetBackupRepository(callCtx, &adminv1.GetBackupRepositoryRequest{RepositoryId: repositoryID})
	if err != nil {
		return nil, err
	}
	return res.GetRepository(), nil
}

func (c *AdminClient) CreateBackupRepository(ctx context.Context, repository *adminv1.BackupRepository) (*adminv1.BackupRepository, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Backup.CreateBackupRepository(callCtx, &adminv1.CreateBackupRepositoryRequest{Repository: repository})
	if err != nil {
		return nil, err
	}
	return res.GetRepository(), nil
}

func (c *AdminClient) UpdateBackupRepository(ctx context.Context, repository *adminv1.BackupRepository) (*adminv1.BackupRepository, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Backup.UpdateBackupRepository(callCtx, &adminv1.UpdateBackupRepositoryRequest{Repository: repository})
	if err != nil {
		return nil, err
	}
	return res.GetRepository(), nil
}

func (c *AdminClient) ValidateBackupRepository(ctx context.Context, repositoryID string) (*adminv1.ValidateBackupRepositoryResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.ValidateBackupRepository(callCtx, &adminv1.ValidateBackupRepositoryRequest{RepositoryId: repositoryID})
}

func (c *AdminClient) DeleteBackupRepository(ctx context.Context, repositoryID string) (*adminv1.DeleteBackupRepositoryResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.DeleteBackupRepository(callCtx, &adminv1.DeleteBackupRepositoryRequest{RepositoryId: repositoryID})
}

func (c *AdminClient) StartBackup(ctx context.Context, req *adminv1.StartBackupRequest) (*adminv1.StartBackupResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	if req == nil {
		req = &adminv1.StartBackupRequest{}
	}
	return c.Backup.StartBackup(callCtx, req)
}

func (c *AdminClient) CancelBackupOperation(ctx context.Context, backupSetID, reason string) (*adminv1.CancelBackupOperationResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.CancelBackupOperation(callCtx, &adminv1.CancelBackupOperationRequest{BackupSetId: backupSetID, Reason: reason})
}

func (c *AdminClient) GetBackupOperation(ctx context.Context, backupSetID string) (*adminv1.GetBackupOperationResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.GetBackupOperation(callCtx, &adminv1.GetBackupOperationRequest{BackupSetId: backupSetID})
}

func (c *AdminClient) ListBackupOperations(ctx context.Context, pageSize int32, pageToken string) (*adminv1.ListBackupOperationsResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.ListBackupOperations(callCtx, &adminv1.ListBackupOperationsRequest{PageSize: pageSize, PageToken: pageToken})
}

func (c *AdminClient) ListBackupSets(ctx context.Context, repositoryID string, pageSize int32, pageToken string) (*adminv1.ListBackupSetsResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.ListBackupSets(callCtx, &adminv1.ListBackupSetsRequest{RepositoryId: repositoryID, PageSize: pageSize, PageToken: pageToken})
}

func (c *AdminClient) GetBackupSet(ctx context.Context, repositoryID, backupSetID string) (*adminv1.BackupSetSummary, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Backup.GetBackupSet(callCtx, &adminv1.GetBackupSetRequest{RepositoryId: repositoryID, BackupSetId: backupSetID})
	if err != nil {
		return nil, err
	}
	return res.GetBackupSet(), nil
}

func (c *AdminClient) ValidateBackupSet(ctx context.Context, repositoryID, backupSetID string, options *adminv1.ValidationOptions) (*adminv1.ValidateBackupSetResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.ValidateBackupSet(callCtx, &adminv1.ValidateBackupSetRequest{RepositoryId: repositoryID, BackupSetId: backupSetID, ValidationOptions: options})
}

func (c *AdminClient) DeleteBackupSet(ctx context.Context, repositoryID, backupSetID, reason string) (*adminv1.DeleteBackupSetResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.DeleteBackupSet(callCtx, &adminv1.DeleteBackupSetRequest{RepositoryId: repositoryID, BackupSetId: backupSetID, Reason: reason})
}

func (c *AdminClient) PlanRepositoryRetention(ctx context.Context, req *adminv1.PlanRepositoryRetentionRequest) (*adminv1.PlanRepositoryRetentionResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	if req == nil {
		req = &adminv1.PlanRepositoryRetentionRequest{}
	}
	return c.Backup.PlanRepositoryRetention(callCtx, req)
}

func (c *AdminClient) ApplyRepositoryRetention(ctx context.Context, req *adminv1.ApplyRepositoryRetentionRequest) (*adminv1.ApplyRepositoryRetentionResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	if req == nil {
		req = &adminv1.ApplyRepositoryRetentionRequest{}
	}
	return c.Backup.ApplyRepositoryRetention(callCtx, req)
}

func (c *AdminClient) GetBackupObservability(ctx context.Context, req *adminv1.GetBackupObservabilityRequest) (*adminv1.GetBackupObservabilityResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	if req == nil {
		req = &adminv1.GetBackupObservabilityRequest{}
	}
	return c.Backup.GetBackupObservability(callCtx, req)
}

func (c *AdminClient) PlanRestore(ctx context.Context, req *adminv1.PlanRestoreRequest) (*adminv1.PlanRestoreResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	if req == nil {
		req = &adminv1.PlanRestoreRequest{}
	}
	return c.Backup.PlanRestore(callCtx, req)
}

func (c *AdminClient) StartRestore(ctx context.Context, req *adminv1.StartRestoreRequest) (*adminv1.StartRestoreResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	if req == nil {
		req = &adminv1.StartRestoreRequest{}
	}
	return c.Backup.StartRestore(callCtx, req)
}

func (c *AdminClient) GetRestoreOperation(ctx context.Context, restoreOperationID string) (*adminv1.GetRestoreOperationResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.GetRestoreOperation(callCtx, &adminv1.GetRestoreOperationRequest{RestoreOperationId: restoreOperationID})
}

func (c *AdminClient) ListRestoreOperations(ctx context.Context, pageSize int32, pageToken string) (*adminv1.ListRestoreOperationsResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.ListRestoreOperations(callCtx, &adminv1.ListRestoreOperationsRequest{PageSize: pageSize, PageToken: pageToken})
}

func (c *AdminClient) CancelRestoreOperation(ctx context.Context, restoreOperationID, reason string) (*adminv1.CancelRestoreOperationResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.CancelRestoreOperation(callCtx, &adminv1.CancelRestoreOperationRequest{RestoreOperationId: restoreOperationID, Reason: reason})
}
