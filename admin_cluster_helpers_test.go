package mycel

import (
	"testing"

	adminv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/admin/v1"
)

func TestClusterDerivedReadinessContractTypesCompile(t *testing.T) {
	readiness := &adminv1.ClusterReadiness{
		ClientReady:          true,
		DerivedReady:         false,
		RestoredClusterReady: false,
		DerivedSubsystems: []*adminv1.ClusterDerivedSubsystemReadiness{{
			Subsystem:       "LEXICAL_SEARCH",
			State:           "rebuild_required",
			RebuildRequired: true,
			Reason:          "derived_artifact_omitted_rebuild_required",
			Evidence:        map[string]string{"backup_set_id": "backup-set-1"},
		}},
	}
	if readiness.GetDerivedReady() || readiness.GetRestoredClusterReady() || !readiness.GetDerivedSubsystems()[0].GetRebuildRequired() {
		t.Fatalf("unexpected cluster derived readiness: %#v", readiness)
	}
}
