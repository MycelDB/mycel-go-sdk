package mycel

import (
	"testing"

	adminv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/admin/v1"
)

func TestBackupRepositoryContractTypesCompile(t *testing.T) {
	repo := &adminv1.BackupRepository{
		RepositoryId:  "repo-prod",
		Provider:      adminv1.BackupRepositoryProvider_BACKUP_REPOSITORY_PROVIDER_S3_COMPATIBLE,
		S3:            &adminv1.S3CompatibleRepositoryConfig{Bucket: "mycel-backups", Prefix: "clusters/prod", Region: "us-east-1"},
		CredentialRef: &adminv1.SecretReference{Scheme: adminv1.SecretReferenceScheme_SECRET_REFERENCE_SCHEME_AWS_IRSA},
		Encryption:    &adminv1.EncryptionPolicy{Mode: adminv1.EncryptionMode_ENCRYPTION_MODE_SSE_KMS, KmsKeyId: "alias/mycel-backups"},
	}
	req := &adminv1.CreateBackupRepositoryRequest{Repository: repo}
	if req.GetRepository().GetProvider() != adminv1.BackupRepositoryProvider_BACKUP_REPOSITORY_PROVIDER_S3_COMPATIBLE {
		t.Fatalf("unexpected repository provider: %#v", req.GetRepository())
	}
}

func TestDeploymentBackupContractTypesCompile(t *testing.T) {
	req := &adminv1.StartBackupRequest{
		RepositoryId:              "repo-prod",
		Reason:                    "before upgrade",
		ArchiveFormat:             adminv1.BackupArchiveFormat_BACKUP_ARCHIVE_FORMAT_TAR_ZST,
		CaptureProvider:           adminv1.BackupCaptureProvider_BACKUP_CAPTURE_PROVIDER_FILESYSTEM_IMMUTABLE_COPY,
		QuiesceReadProfile:        adminv1.BackupQuiesceReadProfile_BACKUP_QUIESCE_READ_PROFILE_CONTROL_ONLY,
		ValidationOptions:         &adminv1.ValidationOptions{RequireCompleteMarker: true},
		ConvergenceTimeoutSeconds: 300,
	}
	res := &adminv1.StartBackupResponse{Operation: &adminv1.BackupOperation{BackupSetId: "backup-set-1", DeploymentMode: adminv1.BackupDeploymentMode_BACKUP_DEPLOYMENT_MODE_STANDALONE, State: adminv1.BackupOperationState_BACKUP_OPERATION_STATE_SUCCEEDED}}
	if req.GetRepositoryId() != "repo-prod" || res.GetOperation().GetDeploymentMode() != adminv1.BackupDeploymentMode_BACKUP_DEPLOYMENT_MODE_STANDALONE {
		t.Fatalf("unexpected deployment backup contract values: %#v %#v", req, res)
	}
}

func TestRestoreContractTypesCompile(t *testing.T) {
	req := &adminv1.PlanRestoreRequest{
		RepositoryId:      "repo-prod",
		BackupSetId:       "backup-set-1",
		TargetMode:        adminv1.RestoreTargetMode_RESTORE_TARGET_MODE_NEW_DEPLOYMENT_BOOTSTRAP,
		IdentityMode:      adminv1.RestoreIdentityMode_RESTORE_IDENTITY_MODE_PRESERVE_BACKUP_ID,
		AuthSessionPolicy: adminv1.RestoreAuthSessionPolicy_RESTORE_AUTH_SESSION_POLICY_PRESERVE_RESTORED_SESSIONS,
		ValidationOptions: &adminv1.ValidationOptions{RequireCompleteMarker: true},
		Target:            &adminv1.RestoreTarget{DeploymentMode: adminv1.BackupDeploymentMode_BACKUP_DEPLOYMENT_MODE_CLUSTERED, ExpectedNodeCount: 3},
	}
	if req.GetTargetMode() != adminv1.RestoreTargetMode_RESTORE_TARGET_MODE_NEW_DEPLOYMENT_BOOTSTRAP || req.GetTarget().GetExpectedNodeCount() != 3 {
		t.Fatalf("unexpected restore target mode: %#v", req)
	}
}

func TestRepositoryRetentionContractTypesCompile(t *testing.T) {
	planReq := &adminv1.PlanRepositoryRetentionRequest{RepositoryId: "repo-prod", RetentionPolicy: &adminv1.RetentionPolicy{KeepLastSuccessful: 3, KeepForSeconds: 86_400, DeleteIncompleteAfterSeconds: 3_600}, Now: "2026-11-22T12:00:00Z"}
	applyReq := &adminv1.ApplyRepositoryRetentionRequest{RepositoryId: "repo-prod", Reason: "test"}
	plan := &adminv1.RepositoryRetentionPlan{RepositoryId: planReq.GetRepositoryId(), DryRun: true, Actions: []*adminv1.RepositoryRetentionAction{{BackupSetId: "backup-set-1", Action: "delete", Reason: "retention_expired"}}}
	if plan.GetRepositoryId() != applyReq.GetRepositoryId() || plan.GetActions()[0].GetAction() != "delete" {
		t.Fatalf("unexpected retention contract values: %#v %#v", planReq, plan)
	}
}

func TestBackupObservabilityContractTypesCompile(t *testing.T) {
	req := &adminv1.GetBackupObservabilityRequest{RepositoryId: "repo-prod", IncludeRetentionProbe: true, StaleOperationAfterSeconds: 3600, BackupFreshnessAfterSeconds: 86400, IncompleteBackupAfterSeconds: 7200}
	snapshot := &adminv1.BackupObservabilitySnapshot{
		ObservedAt:   "2026-11-22T12:00:00Z",
		RepositoryId: req.GetRepositoryId(),
		SummaryState: "warning",
		Metrics:      []*adminv1.BackupMetric{{Name: "mycel_backup_cluster_latest_success_age_seconds", Unit: "seconds", Value: 42, Labels: map[string]string{"repository_id": req.GetRepositoryId()}}},
		Alerts:       []*adminv1.BackupAlertCondition{{Name: "MycelBackupFreshnessStale", Severity: "warning", Firing: true, Reason: "latest_success_stale", Expression: "mycel_backup_cluster_latest_success_age_seconds > 86400"}},
	}
	res := &adminv1.GetBackupObservabilityResponse{Snapshot: snapshot}
	if !req.GetIncludeRetentionProbe() || res.GetSnapshot().GetMetrics()[0].GetValue() != 42 || !res.GetSnapshot().GetAlerts()[0].GetFiring() {
		t.Fatalf("unexpected observability contract values: %#v %#v", req, res)
	}
}

func TestBackupSetValidationSummaryFieldsCompile(t *testing.T) {
	summary := &adminv1.BackupSetSummary{
		BackupSetId:    "backup-set-1",
		DeploymentMode: adminv1.BackupDeploymentMode_BACKUP_DEPLOYMENT_MODE_CLUSTERED,
		Validation:     &adminv1.ValidationSummary{ValidatorVersion: "v1", ExpectedNodeCount: 3, NodeArtifactCount: 3, RequiredArtifactCount: 3, ValidatedArtifactCount: 3, TotalArtifactBytes: 1024, ChecksumAlgorithm: "SHA256", Errors: []string{}},
		Compatibility: &adminv1.CompatibilitySummary{
			RequiredFeatures:                          []string{"checkpoint_barrier_set.v1"},
			StorageFormatVersions:                     map[string]string{"system_raft": "1"},
			MinRestoreMycelVersion:                    "v0.25.0",
			SupportedRestoreTargetModes:               []string{"NEW_DEPLOYMENT_BOOTSTRAP"},
			SupportedIdentityModes:                    []string{"PRESERVE_BACKUP_ID"},
			RequiresSameTopology:                      true,
			RequiresSourceFencingForPreservedIdentity: true,
		},
		Manifest: &adminv1.BackupSetManifestSummary{ManifestVersion: 1, ManifestSizeBytes: 2048, TotalArtifactBytes: 1024, ArtifactCount: 3},
	}
	if summary.GetValidation().GetChecksumAlgorithm() != "SHA256" || !summary.GetCompatibility().GetRequiresSameTopology() || summary.GetManifest().GetArtifactCount() != 3 {
		t.Fatalf("unexpected backup-set validation summary: %#v", summary)
	}
}
