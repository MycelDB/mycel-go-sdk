package mycel

import (
	"testing"

	adminv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/admin/v1"
	"google.golang.org/protobuf/proto"
)

func TestUsageAccountingContractTypesCompile(t *testing.T) {
	scope := &adminv1.UsageScope{Scope: &adminv1.UsageScope_Domain{Domain: &adminv1.DomainUsageScope{SpaceId: "space-1", DomainId: "domain-1"}}}
	revision := &adminv1.UsageRevision{RevisionToken: "usage-42", SourceRevisions: []*adminv1.UsageSourceRevision{{Source: "partition-0", Revision: "raft-index-10"}}}
	req := &adminv1.GetUsageRequest{
		Scope:                     scope,
		DomainInclusion:           adminv1.DomainInclusionMode_DOMAIN_INCLUSION_MODE_RETAINED_NON_SYSTEM,
		RequireQuotaSafe:          true,
		MaxStalenessSeconds:       60,
		MinimumRevision:           revision,
		MetricNames:               []string{"blob.active_logical_bytes"},
		IncludeDiagnosticPhysical: false,
	}
	snapshot := &adminv1.UsageSnapshot{
		Scope:             req.GetScope(),
		DomainInclusion:   req.GetDomainInclusion(),
		AsOfRevision:      revision,
		AccountingState:   adminv1.UsageAccountingState_USAGE_ACCOUNTING_STATE_READY,
		QuotaSafeComplete: true,
		Metrics:           []*adminv1.UsageMetric{{Name: "blob.active_logical_bytes", MetricVersion: 1, Value: 1024, Unit: adminv1.UsageUnit_USAGE_UNIT_BYTES, Authority: adminv1.UsageAuthority_USAGE_AUTHORITY_RAFT_COMMITTED_EXACT, QuotaSafe: true, HardLimitEligible: true, AsOfRevision: revision}},
		Blockers:          []*adminv1.UsageStatus{},
		Warnings:          []*adminv1.UsageStatus{},
	}
	res := &adminv1.GetUsageResponse{Snapshot: snapshot}
	if !req.GetRequireQuotaSafe() || res.GetSnapshot().GetMetrics()[0].GetValue() != 1024 || !res.GetSnapshot().GetQuotaSafeComplete() {
		t.Fatalf("unexpected usage accounting contract values: %#v %#v", req, res)
	}
}

func TestGraphLogicalBytesUsageMetricRoundTrip(t *testing.T) {
	metric := &adminv1.UsageMetric{
		Name:          "graph.logical_bytes",
		MetricVersion: 1,
		Value:         42,
		Unit:          adminv1.UsageUnit_USAGE_UNIT_BYTES,
		Authority:     adminv1.UsageAuthority_USAGE_AUTHORITY_RAFT_COMMITTED_EXACT,
		QuotaSafe:     true,
	}
	encoded, err := proto.Marshal(metric)
	if err != nil {
		t.Fatal(err)
	}
	var decoded adminv1.UsageMetric
	if err := proto.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(metric, &decoded) {
		t.Fatalf("decoded metric = %v, want %v", &decoded, metric)
	}
}

func TestUsageAccountingBatchAndHealthContractTypesCompile(t *testing.T) {
	spaceScope := &adminv1.UsageScope{Scope: &adminv1.UsageScope_Space{Space: &adminv1.SpaceUsageScope{SpaceId: "space-1"}}}
	status := &adminv1.UsageStatus{Reason: adminv1.UsageStatusReason_USAGE_STATUS_REASON_COUNTERS_REBUILDING, Severity: adminv1.UsageStatusSeverity_USAGE_STATUS_SEVERITY_BLOCKER, Scope: spaceScope, MetricName: "blob.active_count", Evidence: map[string]string{"state": "rebuilding"}}
	batch := &adminv1.BatchGetUsageResponse{Results: []*adminv1.BatchUsageResult{{Scope: spaceScope, Result: &adminv1.BatchUsageResult_Status{Status: status}}}}
	health := &adminv1.GetUsageAccountingHealthResponse{Health: &adminv1.UsageAccountingHealth{Scope: spaceScope, State: adminv1.UsageAccountingState_USAGE_ACCOUNTING_STATE_REBUILDING, QuotaSafeReady: false, Blockers: []*adminv1.UsageStatus{status}, RebuildingScopeCount: 1}}
	if batch.GetResults()[0].GetStatus().GetReason() != adminv1.UsageStatusReason_USAGE_STATUS_REASON_COUNTERS_REBUILDING || health.GetHealth().GetQuotaSafeReady() {
		t.Fatalf("unexpected usage health contract values: %#v %#v", batch, health)
	}
}

func TestUsageAdmissionContractTypesCompile(t *testing.T) {
	scope := &adminv1.UsageScope{Scope: &adminv1.UsageScope_Domain{Domain: &adminv1.DomainUsageScope{SpaceId: "space-1", DomainId: "domain-1"}}}
	limit := &adminv1.UsageLimit{Scope: scope, MetricName: "blob.active_logical_bytes", MetricVersion: 1, HardLimit: 1024, Enabled: true, PolicyRef: "plan/example"}
	upsert := &adminv1.UpsertUsageLimitRequest{Limit: limit}
	listed := &adminv1.ListUsageLimitsResponse{Limits: []*adminv1.UsageLimit{limit}, NextPageToken: "next"}
	if !upsert.GetLimit().GetEnabled() || listed.GetLimits()[0].GetHardLimit() != 1024 {
		t.Fatalf("unexpected usage admission contract values: %#v %#v", upsert, listed)
	}
}

func TestUsageAccountingHelperMethodsCompile(t *testing.T) {
	var client *AdminClient
	_ = client.GetUsage
	_ = client.BatchGetUsage
	_ = client.ListUsage
	_ = client.GetUsageAccountingHealth
	_ = client.UpsertUsageLimit
	_ = client.DeleteUsageLimit
	_ = client.GetUsageLimit
	_ = client.ListUsageLimits
}
