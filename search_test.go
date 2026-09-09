package mycel

import (
	"testing"

	clientv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/client/v1"
)

func TestHybridSearchRequestConstructs(t *testing.T) {
	minScore := 0.2
	req := &clientv1.SearchRequest{
		SpaceId:  "space",
		DomainId: "domain",
		Mode:     clientv1.SearchMode_SEARCH_MODE_HYBRID,
		Query:    "raft recovery",
		PageSize: 20,
		Filters: &clientv1.SearchFilters{
			NodeLabels: []string{"Note"},
			Properties: []*clientv1.PropertyFilter{{Path: "tags", Operator: clientv1.FilterOperator_FILTER_OPERATOR_CONTAINS, Values: []string{"k3s"}}},
		},
		Hybrid:   &clientv1.HybridSearchOptions{LexicalWeight: 0.6, SemanticWeight: 0.4, FusionStrategy: clientv1.HybridFusionStrategy_HYBRID_FUSION_STRATEGY_WEIGHTED_RECIPROCAL_RANK},
		Lexical:  &clientv1.LexicalSearchOptions{CandidateCount: 100},
		Semantic: &clientv1.SemanticSearchOptions{MinScore: &minScore, CandidateCount: 100},
	}
	if req.GetMode() != clientv1.SearchMode_SEARCH_MODE_HYBRID {
		t.Fatalf("mode = %v", req.GetMode())
	}
	if req.GetHybrid().GetLexicalWeight() != 0.6 || req.GetHybrid().GetSemanticWeight() != 0.4 {
		t.Fatalf("hybrid weights = %#v", req.GetHybrid())
	}
	if req.GetFilters().GetProperties()[0].GetOperator() != clientv1.FilterOperator_FILTER_OPERATOR_CONTAINS {
		t.Fatalf("property filter = %#v", req.GetFilters().GetProperties()[0])
	}
}
