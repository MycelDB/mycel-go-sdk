package mycel

import (
	"context"

	clientv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/client/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
)

func (c *Client) CreateNode(ctx context.Context, txID string, node *clientv1.NodeCreate) (*clientv1.Node, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Graph.CreateNode(callCtx, &clientv1.CreateNodeRequest{TransactionId: txID, Node: node})
	if err != nil {
		return nil, err
	}
	return res.GetNode(), nil
}

func (c *Client) ApplyGraphOperations(ctx context.Context, txID string, ops []*clientv1.GraphOperation) (*clientv1.ApplyGraphOperationsResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Graph.ApplyGraphOperations(callCtx, &clientv1.ApplyGraphOperationsRequest{TransactionId: txID, Operations: ops})
}

func (c *Client) UpdateNodeContent(ctx context.Context, txID, nodeID, content string) (*clientv1.Node, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	payload, err := structpb.NewStruct(map[string]any{"text": content})
	if err != nil {
		return nil, err
	}
	res, err := c.Graph.UpdateNode(callCtx, &clientv1.UpdateNodeRequest{TransactionId: txID, Node: &clientv1.Node{NodeId: nodeID, Payload: payload}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"payload"}}})
	if err != nil {
		return nil, err
	}
	return res.GetNode(), nil
}

func (c *Client) DeleteNode(ctx context.Context, txID, nodeID string, recursive bool) error {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	_, err := c.Graph.DeleteNode(callCtx, &clientv1.DeleteNodeRequest{TransactionId: txID, NodeId: nodeID, Recursive: recursive})
	return err
}

func (c *Client) CreateEdge(ctx context.Context, txID string, edge *clientv1.EdgeCreate) (*clientv1.Edge, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Graph.CreateEdge(callCtx, &clientv1.CreateEdgeRequest{TransactionId: txID, Edge: edge})
	if err != nil {
		return nil, err
	}
	return res.GetEdge(), nil
}

func (c *Client) GetNode(ctx context.Context, txID, nodeID string) (*clientv1.Node, error) {
	return DoReadValue(ctx, c, "get node", func() (*clientv1.Node, error) {
		callCtx, cancel := c.AuthCallContext(ctx)
		defer cancel()
		res, err := c.Graph.GetNode(callCtx, &clientv1.GetNodeRequest{TransactionId: txID, NodeId: nodeID})
		if err != nil {
			return nil, err
		}
		return res.GetNode(), nil
	})
}

func (c *Client) ListNodes(ctx context.Context, txID string, pageSize int32, pageToken string) (*clientv1.ListNodesResponse, error) {
	return DoReadValue(ctx, c, "list nodes", func() (*clientv1.ListNodesResponse, error) {
		callCtx, cancel := c.AuthCallContext(ctx)
		defer cancel()
		return c.Graph.ListNodes(callCtx, &clientv1.ListNodesRequest{TransactionId: txID, PageSize: pageSize, PageToken: pageToken})
	})
}

func (c *Client) ListChildren(ctx context.Context, txID, parentID string) (*clientv1.ListChildrenResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Graph.ListChildren(callCtx, &clientv1.ListChildrenRequest{TransactionId: txID, ParentNodeId: parentID})
}

func (c *Client) GetParent(ctx context.Context, txID, childID string) (*clientv1.GetParentResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Graph.GetParent(callCtx, &clientv1.GetParentRequest{TransactionId: txID, ChildNodeId: childID})
}

func (c *Client) ExecuteQuery(ctx context.Context, txID string, query *clientv1.GraphQuery, pageSize int32) (*clientv1.ExecuteQueryResponse, error) {
	return DoReadValue(ctx, c, "execute query", func() (*clientv1.ExecuteQueryResponse, error) {
		callCtx, cancel := c.AuthCallContext(ctx)
		defer cancel()
		return c.Query.ExecuteQuery(callCtx, &clientv1.ExecuteQueryRequest{TransactionId: txID, Query: query, PageSize: pageSize})
	})
}

func (c *Client) ExecuteGQL(ctx context.Context, txID string, query string, params map[string]*structpb.Value, pageSize int32) (*clientv1.QueryResult, error) {
	return DoReadValue(ctx, c, "execute gql", func() (*clientv1.QueryResult, error) {
		callCtx, cancel := c.AuthCallContext(ctx)
		defer cancel()
		res, err := c.Query.ExecuteGQL(callCtx, &clientv1.ExecuteGQLRequest{TransactionId: txID, Query: query, Params: params, PageSize: pageSize})
		if err != nil {
			return nil, err
		}
		return res.GetResult(), nil
	})
}

func (c *Client) QueryGQLReadOnly(ctx context.Context, spaceID, domainID, query string, pageSize int32) (*clientv1.QueryResult, error) {
	sessionID, err := c.OpenSession(ctx, spaceID, domainID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = c.CloseSession(ctx, sessionID) }()

	txID, err := c.BeginReadOnlyTransaction(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	closed := false
	defer func() {
		if !closed {
			_ = c.CloseTransaction(ctx, txID)
		}
	}()

	result, err := c.ExecuteGQL(ctx, txID, query, nil, pageSize)
	if err != nil {
		return nil, err
	}
	if err := c.CloseTransaction(ctx, txID); err != nil {
		return nil, err
	}
	closed = true
	return result, nil
}
