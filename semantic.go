package mycel

import (
	"context"

	clientv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/client/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// SemanticSearch executes one direct-semantic search request. Page tokens are
// opaque and the helper never follows or restarts pagination automatically.
func (c *Client) SemanticSearch(ctx context.Context, request *clientv1.SemanticSearchRequest) (*clientv1.SemanticSearchResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Semantic.SemanticSearch(callCtx, request)
}

// NewSemanticEligibilityFilter constructs a conjunctive hard-eligibility
// filter. The returned filter owns copies of the supplied slices.
func NewSemanticEligibilityFilter(nodeLabels, nodeIDs []string, properties ...*clientv1.SemanticPropertyPredicate) *clientv1.SemanticEligibilityFilter {
	return &clientv1.SemanticEligibilityFilter{
		NodeLabels: append([]string(nil), nodeLabels...),
		NodeIds:    append([]string(nil), nodeIDs...),
		Properties: append([]*clientv1.SemanticPropertyPredicate(nil), properties...),
	}
}

// SemanticEqual constructs an exact, type-sensitive equality predicate.
// Passing nil produces an explicit protobuf null value; use SemanticMissing to
// test whether the property is absent.
func SemanticEqual(key string, value any) (*clientv1.SemanticPropertyPredicate, error) {
	return semanticValuePredicate(key, clientv1.SemanticPropertyOperator_SEMANTIC_PROPERTY_OPERATOR_EQUAL, value)
}

// SemanticNotEqual constructs an exact, type-sensitive inequality predicate.
func SemanticNotEqual(key string, value any) (*clientv1.SemanticPropertyPredicate, error) {
	return semanticValuePredicate(key, clientv1.SemanticPropertyOperator_SEMANTIC_PROPERTY_OPERATOR_NOT_EQUAL, value)
}

// SemanticExists matches a present property, including an explicit null.
func SemanticExists(key string) *clientv1.SemanticPropertyPredicate {
	return &clientv1.SemanticPropertyPredicate{Key: key, Operator: clientv1.SemanticPropertyOperator_SEMANTIC_PROPERTY_OPERATOR_EXISTS}
}

// SemanticMissing matches only an absent property.
func SemanticMissing(key string) *clientv1.SemanticPropertyPredicate {
	return &clientv1.SemanticPropertyPredicate{Key: key, Operator: clientv1.SemanticPropertyOperator_SEMANTIC_PROPERTY_OPERATOR_MISSING}
}

func semanticValuePredicate(key string, operator clientv1.SemanticPropertyOperator, value any) (*clientv1.SemanticPropertyPredicate, error) {
	typedValue, err := structpb.NewValue(value)
	if err != nil {
		return nil, err
	}
	return &clientv1.SemanticPropertyPredicate{Key: key, Operator: operator, Value: typedValue}, nil
}
