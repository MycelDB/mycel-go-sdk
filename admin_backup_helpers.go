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

func (c *AdminClient) TriggerBackup(ctx context.Context, reason string) (*adminv1.TriggerBackupResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.TriggerBackup(callCtx, &adminv1.TriggerBackupRequest{Reason: reason})
}

func (c *AdminClient) GetBackupStatus(ctx context.Context) (*adminv1.GetBackupStatusResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.GetBackupStatus(callCtx, &adminv1.GetBackupStatusRequest{})
}

func (c *AdminClient) ListBackups(ctx context.Context, pageSize int32, pageToken string) (*adminv1.ListBackupsResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.ListBackups(callCtx, &adminv1.ListBackupsRequest{PageSize: pageSize, PageToken: pageToken})
}

func (c *AdminClient) DeleteBackup(ctx context.Context, backupID string) (*adminv1.DeleteBackupResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.DeleteBackup(callCtx, &adminv1.DeleteBackupRequest{BackupId: backupID})
}

func (c *AdminClient) StartClusterBackup(ctx context.Context, req *adminv1.StartClusterBackupRequest) (*adminv1.StartClusterBackupResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	if req == nil {
		req = &adminv1.StartClusterBackupRequest{}
	}
	return c.Backup.StartClusterBackup(callCtx, req)
}

func (c *AdminClient) CancelClusterBackup(ctx context.Context, backupSetID, reason string) (*adminv1.CancelClusterBackupResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.CancelClusterBackup(callCtx, &adminv1.CancelClusterBackupRequest{BackupSetId: backupSetID, Reason: reason})
}

func (c *AdminClient) GetClusterBackupStatus(ctx context.Context, backupSetID string) (*adminv1.GetClusterBackupStatusResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.GetClusterBackupStatus(callCtx, &adminv1.GetClusterBackupStatusRequest{BackupSetId: backupSetID})
}

func (c *AdminClient) ListClusterBackups(ctx context.Context, pageSize int32, pageToken string) (*adminv1.ListClusterBackupsResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.ListClusterBackups(callCtx, &adminv1.ListClusterBackupsRequest{PageSize: pageSize, PageToken: pageToken})
}

func (c *AdminClient) ListClusterBackupOperations(ctx context.Context, pageSize int32, pageToken string) (*adminv1.ListClusterBackupOperationsResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.ListClusterBackupOperations(callCtx, &adminv1.ListClusterBackupOperationsRequest{PageSize: pageSize, PageToken: pageToken})
}

func (c *AdminClient) ListClusterBackupSets(ctx context.Context, repositoryID string, pageSize int32, pageToken string) (*adminv1.ListClusterBackupSetsResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.ListClusterBackupSets(callCtx, &adminv1.ListClusterBackupSetsRequest{RepositoryId: repositoryID, PageSize: pageSize, PageToken: pageToken})
}

func (c *AdminClient) GetClusterBackupSet(ctx context.Context, repositoryID, backupSetID string) (*adminv1.ClusterBackupSetSummary, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Backup.GetClusterBackupSet(callCtx, &adminv1.GetClusterBackupSetRequest{RepositoryId: repositoryID, BackupSetId: backupSetID})
	if err != nil {
		return nil, err
	}
	return res.GetBackupSet(), nil
}

func (c *AdminClient) ValidateClusterBackupSet(ctx context.Context, backupSetPath string) (*adminv1.ValidateClusterBackupSetResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.ValidateClusterBackupSet(callCtx, &adminv1.ValidateClusterBackupSetRequest{BackupSetPath: backupSetPath})
}

func (c *AdminClient) ValidateRepositoryClusterBackupSet(ctx context.Context, repositoryID, backupSetID string, options *adminv1.ValidationOptions) (*adminv1.ValidateClusterBackupSetResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.ValidateClusterBackupSet(callCtx, &adminv1.ValidateClusterBackupSetRequest{RepositoryId: repositoryID, BackupSetId: backupSetID, ValidationOptions: options})
}

func (c *AdminClient) DeleteClusterBackupSet(ctx context.Context, repositoryID, backupSetID, reason string) (*adminv1.DeleteClusterBackupSetResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.DeleteClusterBackupSet(callCtx, &adminv1.DeleteClusterBackupSetRequest{RepositoryId: repositoryID, BackupSetId: backupSetID, Reason: reason})
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

func (c *AdminClient) PlanClusterRestore(ctx context.Context, req *adminv1.PlanClusterRestoreRequest) (*adminv1.PlanClusterRestoreResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	if req == nil {
		req = &adminv1.PlanClusterRestoreRequest{}
	}
	return c.Backup.PlanClusterRestore(callCtx, req)
}

func (c *AdminClient) StartClusterRestore(ctx context.Context, req *adminv1.StartClusterRestoreRequest) (*adminv1.StartClusterRestoreResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	if req == nil {
		req = &adminv1.StartClusterRestoreRequest{}
	}
	return c.Backup.StartClusterRestore(callCtx, req)
}

func (c *AdminClient) GetClusterRestoreStatus(ctx context.Context, restoreOperationID string) (*adminv1.GetClusterRestoreStatusResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.GetClusterRestoreStatus(callCtx, &adminv1.GetClusterRestoreStatusRequest{RestoreOperationId: restoreOperationID})
}

func (c *AdminClient) ListClusterRestores(ctx context.Context, pageSize int32, pageToken string) (*adminv1.ListClusterRestoresResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.ListClusterRestores(callCtx, &adminv1.ListClusterRestoresRequest{PageSize: pageSize, PageToken: pageToken})
}

func (c *AdminClient) CancelClusterRestore(ctx context.Context, restoreOperationID, reason string) (*adminv1.CancelClusterRestoreResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.CancelClusterRestore(callCtx, &adminv1.CancelClusterRestoreRequest{RestoreOperationId: restoreOperationID, Reason: reason})
}
