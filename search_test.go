package mycel

import (
	"testing"

	clientv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/client/v1"
	"google.golang.org/protobuf/proto"
)

func TestHybridSearchRequestConstructs(t *testing.T) {
	minScore := 0.2
	req := &clientv1.SearchRequest{
		SpaceId:  "space",
		DomainId: "domain",
		Mode:     clientv1.SearchMode_SEARCH_MODE_HYBRID,
		PageSize: 20,
		Filters: &clientv1.SearchFilters{
			NodeLabels: []string{"Note"},
			Properties: []*clientv1.PropertyFilter{{Path: "tags", Operator: clientv1.FilterOperator_FILTER_OPERATOR_CONTAINS, Values: []string{"k3s"}}},
		},
		Hybrid:   &clientv1.HybridSearchOptions{LexicalWeight: 0.6, SemanticWeight: 0.4, FusionStrategy: clientv1.HybridFusionStrategy_HYBRID_FUSION_STRATEGY_WEIGHTED_RECIPROCAL_RANK},
		Lexical:  &clientv1.LexicalSearchOptions{CandidateCount: 100, Query: proto.String(`title:"raft recovery"`)},
		Semantic: &clientv1.SemanticSearchOptions{MinScore: &minScore, CandidateCount: 100, Query: proto.String("how raft recovery works")},
	}
	if req.GetMode() != clientv1.SearchMode_SEARCH_MODE_HYBRID {
		t.Fatalf("mode = %v", req.GetMode())
	}
	if req.GetHybrid().GetLexicalWeight() != 0.6 || req.GetHybrid().GetSemanticWeight() != 0.4 {
		t.Fatalf("hybrid weights = %#v", req.GetHybrid())
	}
	if req.GetLexical().GetQuery() != `title:"raft recovery"` || req.GetSemantic().GetQuery() != "how raft recovery works" {
		t.Fatalf("channel queries = lexical %q semantic %q", req.GetLexical().GetQuery(), req.GetSemantic().GetQuery())
	}
	if req.GetFilters().GetProperties()[0].GetOperator() != clientv1.FilterOperator_FILTER_OPERATOR_CONTAINS {
		t.Fatalf("property filter = %#v", req.GetFilters().GetProperties()[0])
	}
}
