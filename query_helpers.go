package mycel

import (
	"context"

	clientv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/client/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func (c *Client) ExplainQuery(ctx context.Context, txID string, query *clientv1.GraphQuery) (*clientv1.QueryDiagnostics, error) {
	return DoReadValue(ctx, c, "explain query", func() (*clientv1.QueryDiagnostics, error) {
		callCtx, cancel := c.AuthCallContext(ctx)
		defer cancel()
		res, err := c.Query.ExplainQuery(callCtx, &clientv1.ExplainQueryRequest{TransactionId: txID, Query: query})
		if err != nil {
			return nil, err
		}
		return res.GetDiagnostics(), nil
	})
}

func (c *Client) ExplainGQL(ctx context.Context, txID string, query string, params map[string]*structpb.Value) (*clientv1.QueryDiagnostics, error) {
	return DoReadValue(ctx, c, "explain gql", func() (*clientv1.QueryDiagnostics, error) {
		callCtx, cancel := c.AuthCallContext(ctx)
		defer cancel()
		res, err := c.Query.ExplainGQL(callCtx, &clientv1.ExplainGQLRequest{TransactionId: txID, Query: query, Params: params})
		if err != nil {
			return nil, err
		}
		return res.GetDiagnostics(), nil
	})
}

func IndexedNodeLookupQuery(alias string, label string, property string, value any, outputName string) (*clientv1.GraphQuery, error) {
	literal, err := structpb.NewValue(value)
	if err != nil {
		return nil, err
	}
	if outputName == "" {
		outputName = alias
	}
	return &clientv1.GraphQuery{
		Match:   &clientv1.GraphPattern{Start: &clientv1.NodePattern{Alias: alias, Labels: []string{label}}},
		Where:   &clientv1.Expr{Expr: &clientv1.Expr_PropertyEquals{PropertyEquals: &clientv1.PropertyEqualsExpr{Alias: alias, Name: property, Value: literal}}},
		Returns: []*clientv1.ReturnProjection{{Alias: alias, OutputName: outputName, Kind: clientv1.ReturnProjectionKind_RETURN_PROJECTION_KIND_NODE}},
	}, nil
}

func OrderedNodeQuery(alias string, label string, property string, direction clientv1.SortDirection, limit int32) *clientv1.GraphQuery {
	if direction == clientv1.SortDirection_SORT_DIRECTION_UNSPECIFIED {
		direction = clientv1.SortDirection_SORT_DIRECTION_ASC
	}
	return &clientv1.GraphQuery{
		Match:   &clientv1.GraphPattern{Start: &clientv1.NodePattern{Alias: alias, Labels: []string{label}}},
		Returns: []*clientv1.ReturnProjection{{Alias: alias, OutputName: alias, Kind: clientv1.ReturnProjectionKind_RETURN_PROJECTION_KIND_NODE}},
		OrderBy: []*clientv1.OrderSpec{{Value: PropValue(alias, property), Direction: direction}},
		Limit:   limit,
	}
}

func TextPredicateQuery(alias string, label string, property string, text string, outputName string) *clientv1.GraphQuery {
	if outputName == "" {
		outputName = alias
	}
	return &clientv1.GraphQuery{
		Match:   &clientv1.GraphPattern{Start: &clientv1.NodePattern{Alias: alias, Labels: []string{label}}},
		Where:   &clientv1.Expr{Expr: &clientv1.Expr_Text{Text: &clientv1.TextSearchExpr{Alias: alias, Field: property, Query: text}}},
		Returns: []*clientv1.ReturnProjection{{Alias: alias, OutputName: outputName, Kind: clientv1.ReturnProjectionKind_RETURN_PROJECTION_KIND_NODE}},
	}
}

func SemanticPredicateQuery(alias string, label string, text string, topK int32, outputName string) *clientv1.GraphQuery {
	if outputName == "" {
		outputName = alias
	}
	return &clientv1.GraphQuery{
		Match:   &clientv1.GraphPattern{Start: &clientv1.NodePattern{Alias: alias, Labels: []string{label}}},
		Where:   &clientv1.Expr{Expr: &clientv1.Expr_Semantic{Semantic: &clientv1.SemanticSearchExpr{Alias: alias, Query: text, Limit: topK}}},
		Returns: []*clientv1.ReturnProjection{{Alias: alias, OutputName: outputName, Kind: clientv1.ReturnProjectionKind_RETURN_PROJECTION_KIND_NODE}},
	}
}

func PathQuery(pathAlias string, startAlias string, startLabel string, edgeKind string, targetAlias string, minDepth int32, maxDepth int32) *clientv1.GraphQuery {
	return &clientv1.GraphQuery{
		Match: &clientv1.GraphPattern{
			Start: &clientv1.NodePattern{Alias: startAlias, Labels: []string{startLabel}},
			Steps: []*clientv1.TraversalStep{{Direction: clientv1.TraversalDirection_TRAVERSAL_DIRECTION_OUT, EdgeKind: edgeKind, Depth: &clientv1.DepthSpec{MinDepth: minDepth, MaxDepth: maxDepth}, Target: &clientv1.NodePattern{Alias: targetAlias}}},
		},
		PathAlias: pathAlias,
		Returns:   []*clientv1.ReturnProjection{{Alias: pathAlias, OutputName: pathAlias, Kind: clientv1.ReturnProjectionKind_RETURN_PROJECTION_KIND_PATH}},
	}
}

func AggregateCountQuery(alias string, label string, outputName string) *clientv1.GraphQuery {
	if outputName == "" {
		outputName = "count"
	}
	return &clientv1.GraphQuery{
		Match:            &clientv1.GraphPattern{Start: &clientv1.NodePattern{Alias: alias, Labels: []string{label}}},
		AggregateReturns: []*clientv1.AggregateProjection{{OutputName: outputName, Function: clientv1.AggregateFunction_AGGREGATE_FUNCTION_COUNT, Argument: AggregateStar()}},
	}
}

func AggregatePropertyQuery(alias string, label string, property string, function clientv1.AggregateFunction, outputName string) *clientv1.GraphQuery {
	return &clientv1.GraphQuery{
		Match:            &clientv1.GraphPattern{Start: &clientv1.NodePattern{Alias: alias, Labels: []string{label}}},
		AggregateReturns: []*clientv1.AggregateProjection{{OutputName: outputName, Function: function, Argument: AggregateValue(PropValue(alias, property))}},
	}
}

func PropValue(alias string, property string) *clientv1.ValueExpr {
	return &clientv1.ValueExpr{Expr: &clientv1.ValueExpr_Prop{Prop: &clientv1.PropExpr{Alias: alias, Name: property}}}
}

func AggregateStar() *clientv1.AggregateArgument {
	return &clientv1.AggregateArgument{Argument: &clientv1.AggregateArgument_Star{Star: true}}
}

func AggregateValue(value *clientv1.ValueExpr) *clientv1.AggregateArgument {
	return &clientv1.AggregateArgument{Argument: &clientv1.AggregateArgument_Value{Value: value}}
}
