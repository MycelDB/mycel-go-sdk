package mycel

import (
	"testing"

	clientv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/client/v1"
)

func TestQueryHelpersBuildCommonShapes(t *testing.T) {
	lookup, err := IndexedNodeLookupQuery("n", "Note", "title", "A", "node")
	if err != nil {
		t.Fatalf("IndexedNodeLookupQuery() error = %v", err)
	}
	if lookup.GetMatch().GetStart().GetAlias() != "n" || lookup.GetWhere().GetPropertyEquals().GetName() != "title" || lookup.GetReturns()[0].GetOutputName() != "node" {
		t.Fatalf("lookup query = %+v", lookup)
	}

	ordered := OrderedNodeQuery("j", "JournalEntry", "date", clientv1.SortDirection_SORT_DIRECTION_DESC, 10)
	if ordered.GetOrderBy()[0].GetValue().GetProp().GetName() != "date" || ordered.GetOrderBy()[0].GetDirection() != clientv1.SortDirection_SORT_DIRECTION_DESC || ordered.GetLimit() != 10 {
		t.Fatalf("ordered query = %+v", ordered)
	}

	text := TextPredicateQuery("d", "Document", "body", "memory", "doc")
	if text.GetWhere().GetText().GetQuery() != "memory" || text.GetWhere().GetText().GetField() != "body" {
		t.Fatalf("text query = %+v", text)
	}

	semantic := SemanticPredicateQuery("d", "Document", "memory", 5, "doc")
	if semantic.GetWhere().GetSemantic().GetLimit() != 5 || semantic.GetWhere().GetSemantic().GetQuery() != "memory" {
		t.Fatalf("semantic query = %+v", semantic)
	}

	path := PathQuery("p", "a", "Note", "REFERENCES", "b", 1, 3)
	if path.GetPathAlias() != "p" || path.GetReturns()[0].GetKind() != clientv1.ReturnProjectionKind_RETURN_PROJECTION_KIND_PATH || path.GetMatch().GetSteps()[0].GetDepth().GetMaxDepth() != 3 {
		t.Fatalf("path query = %+v", path)
	}

	count := AggregateCountQuery("n", "Note", "total")
	if count.GetAggregateReturns()[0].GetFunction() != clientv1.AggregateFunction_AGGREGATE_FUNCTION_COUNT || !count.GetAggregateReturns()[0].GetArgument().GetStar() {
		t.Fatalf("count query = %+v", count)
	}

	avg := AggregatePropertyQuery("n", "Note", "score", clientv1.AggregateFunction_AGGREGATE_FUNCTION_AVG, "avg_score")
	if avg.GetAggregateReturns()[0].GetArgument().GetValue().GetProp().GetName() != "score" || avg.GetAggregateReturns()[0].GetFunction() != clientv1.AggregateFunction_AGGREGATE_FUNCTION_AVG {
		t.Fatalf("avg query = %+v", avg)
	}
}
