package mycel

import (
	"context"
	"regexp"
	"strings"
	"testing"

	clientv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/client/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func TestBeginAndCommitIdempotentTransactionPreserveOptions(t *testing.T) {
	transactions := &fakeTransactionServiceClient{
		beginResponse:  &clientv1.BeginTransactionResponse{Transaction: &clientv1.GraphTransaction{TransactionId: "tx-1"}},
		commitResponse: &clientv1.CommitTransactionResponse{Commit: &clientv1.TransactionCommit{CommitId: "commit-1"}},
	}
	c := &Client{Transaction: transactions}
	tx, err := c.BeginIdempotentReadWriteTransaction(context.Background(), "session-1", "space-1", "domain-1", "capture-1")
	if err != nil {
		t.Fatal(err)
	}
	if tx.Transaction.GetTransactionId() != "tx-1" || tx.CommitOptions.IdempotencyKey != "capture-1" {
		t.Fatalf("idempotent transaction = %#v", tx)
	}
	if _, err := c.CommitIdempotentTransaction(context.Background(), tx); err != nil {
		t.Fatal(err)
	}
	if request := transactions.commitRequest; request.GetIdempotencyKey() != "capture-1" || request.GetSpaceId() != "space-1" || request.GetDomainId() != "domain-1" {
		t.Fatalf("commit request = %#v", request)
	}
}

func TestCommitIdempotentTransactionWithRetryPreservesRequest(t *testing.T) {
	transactions := &fakeTransactionServiceClient{
		commitErrors:   []error{status.Error(codes.Unavailable, "lost response"), nil},
		commitResponse: &clientv1.CommitTransactionResponse{Commit: &clientv1.TransactionCommit{CommitId: "commit-1"}, IdempotentReplay: true},
	}
	c := &Client{Transaction: transactions}
	tx := &IdempotentTransaction{Transaction: &clientv1.GraphTransaction{TransactionId: "tx-1"}, CommitOptions: CommitTransactionOptions{IdempotencyKey: "capture-1", SpaceID: "space-1", DomainID: "domain-1"}}
	res, err := c.CommitIdempotentTransactionWithRetry(context.Background(), tx, IdempotentCommitRetryPolicy{MaxAttempts: 3})
	if err != nil {
		t.Fatal(err)
	}
	if transactions.commitCalls != 2 || !res.GetIdempotentReplay() {
		t.Fatalf("commit calls = %d, response = %#v", transactions.commitCalls, res)
	}
	for _, request := range transactions.commitRequests {
		if request.GetTransactionId() != "tx-1" || request.GetIdempotencyKey() != "capture-1" || request.GetSpaceId() != "space-1" || request.GetDomainId() != "domain-1" {
			t.Fatalf("retry changed request: %#v", request)
		}
	}
}

func TestCommitIdempotentTransactionWithRetryDoesNotRetryConflict(t *testing.T) {
	transactions := &fakeTransactionServiceClient{commitErrors: []error{status.Error(codes.AlreadyExists, "key mismatch")}}
	c := &Client{Transaction: transactions}
	tx := &IdempotentTransaction{Transaction: &clientv1.GraphTransaction{TransactionId: "tx-1"}, CommitOptions: CommitTransactionOptions{IdempotencyKey: "capture-1", SpaceID: "space-1", DomainID: "domain-1"}}
	if _, err := c.CommitIdempotentTransactionWithRetry(context.Background(), tx, IdempotentCommitRetryPolicy{MaxAttempts: 3}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("commit error = %v", err)
	}
	if transactions.commitCalls != 1 {
		t.Fatalf("commit calls = %d, want 1", transactions.commitCalls)
	}
}

func TestCommitTransactionIdempotentSendsStableScopeAndReturnsReplay(t *testing.T) {
	transactions := &fakeTransactionServiceClient{commitResponse: &clientv1.CommitTransactionResponse{
		Commit:           &clientv1.TransactionCommit{CommitId: "commit-1", IdempotencyKey: "capture-1"},
		IdempotentReplay: true,
	}}
	c := &Client{Transaction: transactions}
	res, err := c.CommitTransactionIdempotent(context.Background(), "tx-1", "space-1", "domain-1", "capture-1")
	if err != nil {
		t.Fatalf("CommitTransactionIdempotent() error = %v", err)
	}
	if !res.GetIdempotentReplay() || res.GetCommit().GetCommitId() != "commit-1" {
		t.Fatalf("unexpected response: %#v", res)
	}
	request := transactions.commitRequest
	if request.GetTransactionId() != "tx-1" || request.GetIdempotencyKey() != "capture-1" || request.GetSpaceId() != "space-1" || request.GetDomainId() != "domain-1" {
		t.Fatalf("unexpected commit request: %#v", request)
	}
}

func TestCommitTransactionWithOptionsRejectsInvalidIdempotencyWithoutRPC(t *testing.T) {
	tests := []CommitTransactionOptions{
		{SpaceID: "space-1"},
		{IdempotencyKey: " key ", SpaceID: "space-1", DomainID: "domain-1"},
		{IdempotencyKey: "key"},
		{IdempotencyKey: strings.Repeat("x", 129), SpaceID: "space-1", DomainID: "domain-1"},
	}
	for _, options := range tests {
		transactions := &fakeTransactionServiceClient{}
		c := &Client{Transaction: transactions}
		if _, err := c.CommitTransactionWithOptions(context.Background(), "tx-1", options); err == nil {
			t.Fatalf("CommitTransactionWithOptions(%#v) succeeded", options)
		}
		if transactions.commitRequest != nil {
			t.Fatalf("invalid options invoked RPC: %#v", transactions.commitRequest)
		}
	}
}

type fakeTransactionServiceClient struct {
	beginRequest   *clientv1.BeginTransactionRequest
	beginResponse  *clientv1.BeginTransactionResponse
	beginErr       error
	commitRequest  *clientv1.CommitTransactionRequest
	commitRequests []*clientv1.CommitTransactionRequest
	commitResponse *clientv1.CommitTransactionResponse
	commitErr      error
	commitErrors   []error
	commitCalls    int
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
	f.commitRequests = append(f.commitRequests, in)
	f.commitCalls++
	if len(f.commitErrors) >= f.commitCalls && f.commitErrors[f.commitCalls-1] != nil {
		return nil, f.commitErrors[f.commitCalls-1]
	}
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
