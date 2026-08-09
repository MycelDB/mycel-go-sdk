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

func (c *AdminClient) TriggerClusterBackup(ctx context.Context, reason, outputDir string, archiveFormat adminv1.BackupArchiveFormat) (*adminv1.TriggerClusterBackupResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.TriggerClusterBackup(callCtx, &adminv1.TriggerClusterBackupRequest{Reason: reason, OutputDir: outputDir, ArchiveFormat: archiveFormat})
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

func (c *AdminClient) ValidateClusterBackupSet(ctx context.Context, backupSetPath string) (*adminv1.ValidateClusterBackupSetResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Backup.ValidateClusterBackupSet(callCtx, &adminv1.ValidateClusterBackupSetRequest{BackupSetPath: backupSetPath})
}
