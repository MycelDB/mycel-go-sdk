package mycel

import (
	"context"
	"regexp"
	"testing"

	clientv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/client/v1"
	"google.golang.org/grpc"
)

func TestNewOperationIDReturnsUUIDV4(t *testing.T) {
	operationID := NewOperationID()
	match, err := regexp.MatchString(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`, operationID)
	if err != nil {
		t.Fatalf("regexp error: %v", err)
	}
	if !match {
		t.Fatalf("operation ID %q is not a UUID v4", operationID)
	}
}

func TestBeginTransactionWithOperationIDSendsOperationID(t *testing.T) {
	operationID := NewOperationID()
	transactions := &fakeTransactionServiceClient{
		beginResponse: &clientv1.BeginTransactionResponse{Transaction: &clientv1.GraphTransaction{TransactionId: "tx-1", OperationId: operationID}},
	}
	c := &Client{Transaction: transactions}

	tx, err := c.BeginReadWriteTransactionWithOperationID(context.Background(), "session-1", operationID)
	if err != nil {
		t.Fatalf("BeginReadWriteTransactionWithOperationID() error = %v", err)
	}
	if tx.GetTransactionId() != "tx-1" || tx.GetOperationId() != operationID {
		t.Fatalf("unexpected transaction: %#v", tx)
	}
	if got := transactions.beginRequest.GetOperationId(); got != operationID {
		t.Fatalf("request operation_id = %q, want %q", got, operationID)
	}
	if got := transactions.beginRequest.GetMode(); got != clientv1.TransactionMode_TRANSACTION_MODE_READ_WRITE {
		t.Fatalf("request mode = %s, want READ_WRITE", got)
	}
}

func TestCommitTransactionResultReturnsOperationID(t *testing.T) {
	operationID := NewOperationID()
	transactions := &fakeTransactionServiceClient{
		commitResponse: &clientv1.CommitTransactionResponse{Commit: &clientv1.TransactionCommit{TransactionId: "tx-1", OperationId: operationID}},
	}
	c := &Client{Transaction: transactions}

	commit, err := c.CommitTransactionResult(context.Background(), "tx-1")
	if err != nil {
		t.Fatalf("CommitTransactionResult() error = %v", err)
	}
	if commit.GetTransactionId() != "tx-1" || commit.GetOperationId() != operationID {
		t.Fatalf("unexpected commit: %#v", commit)
	}
	if got := transactions.commitRequest.GetTransactionId(); got != "tx-1" {
		t.Fatalf("request transaction_id = %q, want tx-1", got)
	}
}

type fakeTransactionServiceClient struct {
	beginRequest   *clientv1.BeginTransactionRequest
	beginResponse  *clientv1.BeginTransactionResponse
	beginErr       error
	commitRequest  *clientv1.CommitTransactionRequest
	commitResponse *clientv1.CommitTransactionResponse
	commitErr      error
}

func (f *fakeTransactionServiceClient) BeginTransaction(ctx context.Context, in *clientv1.BeginTransactionRequest, opts ...grpc.CallOption) (*clientv1.BeginTransactionResponse, error) {
	f.beginRequest = in
	if f.beginErr != nil {
		return nil, f.beginErr
	}
	return f.beginResponse, nil
}

func (f *fakeTransactionServiceClient) GetTransaction(ctx context.Context, in *clientv1.GetTransactionRequest, opts ...grpc.CallOption) (*clientv1.GetTransactionResponse, error) {
	return nil, nil
}

func (f *fakeTransactionServiceClient) CommitTransaction(ctx context.Context, in *clientv1.CommitTransactionRequest, opts ...grpc.CallOption) (*clientv1.CommitTransactionResponse, error) {
	f.commitRequest = in
	if f.commitErr != nil {
		return nil, f.commitErr
	}
	return f.commitResponse, nil
}

func (f *fakeTransactionServiceClient) RollbackTransaction(ctx context.Context, in *clientv1.RollbackTransactionRequest, opts ...grpc.CallOption) (*clientv1.RollbackTransactionResponse, error) {
	return nil, nil
}

func (f *fakeTransactionServiceClient) CloseTransaction(ctx context.Context, in *clientv1.CloseTransactionRequest, opts ...grpc.CallOption) (*clientv1.CloseTransactionResponse, error) {
	return nil, nil
}
