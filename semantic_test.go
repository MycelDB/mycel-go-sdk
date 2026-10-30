package mycel

import (
	"context"
	"testing"

	clientv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/client/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestSemanticPredicateConstructorsPreserveNullAndMissing(t *testing.T) {
	equal, err := SemanticEqual("archived_at", nil)
	if err != nil {
		t.Fatal(err)
	}
	if equal.GetOperator() != clientv1.SemanticPropertyOperator_SEMANTIC_PROPERTY_OPERATOR_EQUAL || equal.GetValue().GetKind() == nil {
		t.Fatalf("equal predicate = %#v", equal)
	}
	if _, ok := equal.GetValue().GetKind().(*structpb.Value_NullValue); !ok {
		t.Fatalf("equal null kind = %T", equal.GetValue().GetKind())
	}

	notEqual, err := SemanticNotEqual("status", "deleted")
	if err != nil {
		t.Fatal(err)
	}
	if notEqual.GetOperator() != clientv1.SemanticPropertyOperator_SEMANTIC_PROPERTY_OPERATOR_NOT_EQUAL || notEqual.GetValue().GetStringValue() != "deleted" {
		t.Fatalf("not-equal predicate = %#v", notEqual)
	}
	if exists := SemanticExists("summary"); exists.GetOperator() != clientv1.SemanticPropertyOperator_SEMANTIC_PROPERTY_OPERATOR_EXISTS || exists.GetValue() != nil {
		t.Fatalf("exists predicate = %#v", exists)
	}
	if missing := SemanticMissing("deleted_at"); missing.GetOperator() != clientv1.SemanticPropertyOperator_SEMANTIC_PROPERTY_OPERATOR_MISSING || missing.GetValue() != nil {
		t.Fatalf("missing predicate = %#v", missing)
	}
}

func TestNewSemanticEligibilityFilterOwnsInputs(t *testing.T) {
	labels := []string{"Article"}
	nodeIDs := []string{"node-1"}
	predicate := SemanticMissing("deleted_at")
	filter := NewSemanticEligibilityFilter(labels, nodeIDs, predicate)
	labels[0], nodeIDs[0] = "Changed", "changed"
	if filter.GetNodeLabels()[0] != "Article" || filter.GetNodeIds()[0] != "node-1" || filter.GetProperties()[0] != predicate {
		t.Fatalf("filter = %#v", filter)
	}
}

func TestClientSemanticSearchForwardsExactRequestAndResponse(t *testing.T) {
	pageToken := "opaque-page-token"
	request := &clientv1.SemanticSearchRequest{
		SpaceId: "space-1", DomainId: "domain-1", Query: "raft recovery", Limit: 25,
		PageToken: &pageToken,
		Filter:    NewSemanticEligibilityFilter([]string{"Article"}, []string{"node-1"}, SemanticExists("summary")),
	}
	response := &clientv1.SemanticSearchResponse{NextPageToken: "opaque-next", FilterExecutionMode: "semantic_index_pushdown", ExaminedSemanticCandidates: 12, EligibleSemanticCandidates: 4}
	fake := &fakeSemanticServiceClient{response: response}
	client := &Client{Semantic: fake}
	got, err := client.SemanticSearch(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(fake.request, request) || got != response || got.GetNextPageToken() != "opaque-next" || got.GetExaminedSemanticCandidates() != 12 || got.GetEligibleSemanticCandidates() != 4 {
		t.Fatalf("request=%#v response=%#v", fake.request, got)
	}
}

func TestClientSemanticSearchPreservesStaleTokenFailure(t *testing.T) {
	fake := &fakeSemanticServiceClient{err: status.Error(codes.FailedPrecondition, "stale continuation")}
	client := &Client{Semantic: fake}
	_, err := client.SemanticSearch(context.Background(), &clientv1.SemanticSearchRequest{})
	if status.Code(err) != codes.FailedPrecondition || fake.calls != 1 {
		t.Fatalf("error=%v calls=%d", err, fake.calls)
	}
}

type fakeSemanticServiceClient struct {
	request  *clientv1.SemanticSearchRequest
	response *clientv1.SemanticSearchResponse
	err      error
	calls    int
}

func (f *fakeSemanticServiceClient) ListSemanticRules(context.Context, *clientv1.ListSemanticRulesRequest, ...grpc.CallOption) (*clientv1.ListSemanticRulesResponse, error) {
	return &clientv1.ListSemanticRulesResponse{}, nil
}

func (f *fakeSemanticServiceClient) SemanticSearch(_ context.Context, request *clientv1.SemanticSearchRequest, _ ...grpc.CallOption) (*clientv1.SemanticSearchResponse, error) {
	f.calls++
	f.request = request
	return f.response, f.err
}
