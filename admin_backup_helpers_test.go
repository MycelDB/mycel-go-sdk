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

func TestRestoreContractTypesCompile(t *testing.T) {
	req := &adminv1.PlanClusterRestoreRequest{
		RepositoryId:      "repo-prod",
		BackupSetId:       "backup-set-1",
		TargetMode:        adminv1.ClusterRestoreTargetMode_CLUSTER_RESTORE_TARGET_MODE_NEW_CLUSTER_BOOTSTRAP,
		IdentityMode:      adminv1.ClusterRestoreIdentityMode_CLUSTER_RESTORE_IDENTITY_MODE_PRESERVE_BACKUP_ID,
		AuthSessionPolicy: adminv1.ClusterRestoreAuthSessionPolicy_CLUSTER_RESTORE_AUTH_SESSION_POLICY_PRESERVE_RESTORED_SESSIONS,
		ValidationOptions: &adminv1.ValidationOptions{RequireCompleteMarker: true},
	}
	if req.GetTargetMode() != adminv1.ClusterRestoreTargetMode_CLUSTER_RESTORE_TARGET_MODE_NEW_CLUSTER_BOOTSTRAP {
		t.Fatalf("unexpected restore target mode: %#v", req)
	}
}

func TestBackupSetValidationSummaryFieldsCompile(t *testing.T) {
	summary := &adminv1.ClusterBackupSetSummary{
		BackupSetId: "backup-set-1",
		Validation:  &adminv1.ValidationSummary{ValidatorVersion: "v1", ExpectedNodeCount: 3, NodeArtifactCount: 3, RequiredArtifactCount: 3, ValidatedArtifactCount: 3, TotalArtifactBytes: 1024, ChecksumAlgorithm: "SHA256", Errors: []string{}},
		Compatibility: &adminv1.CompatibilitySummary{
			RequiredFeatures:                          []string{"checkpoint_barrier_set.v1"},
			StorageFormatVersions:                     map[string]string{"system_raft": "1"},
			MinRestoreMycelVersion:                    "v0.24.0",
			SupportedRestoreTargetModes:               []string{"NEW_CLUSTER_BOOTSTRAP"},
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
