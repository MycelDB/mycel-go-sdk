package mycel

import (
	"context"
	"testing"

	clientv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/client/v1"
	"google.golang.org/grpc"
)

type fakeAutomationServiceClient struct {
	clientv1.AutomationServiceClient
	startRequest  *clientv1.StartGraphAutomationReconciliationRequest
	getRequest    *clientv1.GetGraphAutomationReconciliationRequest
	cancelRequest *clientv1.CancelGraphAutomationReconciliationRequest
	operation     *clientv1.GraphAutomationReconciliation
}

func (f *fakeAutomationServiceClient) StartGraphAutomationReconciliation(_ context.Context, request *clientv1.StartGraphAutomationReconciliationRequest, _ ...grpc.CallOption) (*clientv1.StartGraphAutomationReconciliationResponse, error) {
	f.startRequest = request
	return &clientv1.StartGraphAutomationReconciliationResponse{Operation: f.operation}, nil
}

func (f *fakeAutomationServiceClient) GetGraphAutomationReconciliation(_ context.Context, request *clientv1.GetGraphAutomationReconciliationRequest, _ ...grpc.CallOption) (*clientv1.GetGraphAutomationReconciliationResponse, error) {
	f.getRequest = request
	return &clientv1.GetGraphAutomationReconciliationResponse{Operation: f.operation}, nil
}

func (f *fakeAutomationServiceClient) CancelGraphAutomationReconciliation(_ context.Context, request *clientv1.CancelGraphAutomationReconciliationRequest, _ ...grpc.CallOption) (*clientv1.CancelGraphAutomationReconciliationResponse, error) {
	f.cancelRequest = request
	return &clientv1.CancelGraphAutomationReconciliationResponse{Operation: f.operation}, nil
}

func TestGraphAutomationReconciliationHelpersForwardRequests(t *testing.T) {
	operation := &clientv1.GraphAutomationReconciliation{OperationId: "operation-1", Status: "queued"}
	service := &fakeAutomationServiceClient{operation: operation}
	client := &Client{Automation: service}

	started, err := client.StartGraphAutomationReconciliation(context.Background(), "space-1", "domain-1", "binding-1", "reconcile-1", 25, 2)
	if err != nil {
		t.Fatal(err)
	}
	if started.GetOperationId() != operation.GetOperationId() || service.startRequest.GetSpaceId() != "space-1" || service.startRequest.GetDomainId() != "domain-1" || service.startRequest.GetBindingId() != "binding-1" || service.startRequest.GetIdempotencyKey() != "reconcile-1" || service.startRequest.GetBatchSize() != 25 || service.startRequest.GetMaxConcurrency() != 2 {
		t.Fatalf("start request=%v operation=%v", service.startRequest, started)
	}

	got, err := client.GetGraphAutomationReconciliation(context.Background(), "space-1", "operation-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.GetOperationId() != operation.GetOperationId() || service.getRequest.GetSpaceId() != "space-1" || service.getRequest.GetOperationId() != "operation-1" {
		t.Fatalf("get request=%v operation=%v", service.getRequest, got)
	}

	canceled, err := client.CancelGraphAutomationReconciliation(context.Background(), "space-1", "operation-1")
	if err != nil {
		t.Fatal(err)
	}
	if canceled.GetOperationId() != operation.GetOperationId() || service.cancelRequest.GetSpaceId() != "space-1" || service.cancelRequest.GetOperationId() != "operation-1" {
		t.Fatalf("cancel request=%v operation=%v", service.cancelRequest, canceled)
	}
}
