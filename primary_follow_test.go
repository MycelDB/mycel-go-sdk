package mycel

import (
	"errors"
	"strings"
	"testing"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func notPrimaryTestError(t *testing.T) error {
	t.Helper()
	st := status.New(codes.FailedPrecondition, "node is not cluster primary")
	withDetails, err := st.WithDetails(&errdetails.ErrorInfo{Reason: NotPrimaryReason, Domain: "mycel.cluster", Metadata: map[string]string{PrimaryNodeIDKey: "node_b", PrimaryNodeNameKey: "node-b", PrimaryBackendKey: "127.0.0.1:9094", AuthorityEpochKey: "2"}})
	if err != nil {
		t.Fatal(err)
	}
	return withDetails.Err()
}

func TestPrimaryHintFromErrorForSwitchover(t *testing.T) {
	hint, ok := PrimaryHintFromError(notPrimaryTestError(t))
	if !ok {
		t.Fatal("hint missing")
	}
	if hint.NodeName != "node-b" || hint.BackendAdvertiseAddr != "127.0.0.1:9094" || hint.AuthorityEpoch != "2" {
		t.Fatalf("hint=%#v", hint)
	}
}

func TestPrimaryChangedRetryRequiredError(t *testing.T) {
	err := &PrimaryChangedRetryRequiredError{Hint: PrimaryHint{BackendAdvertiseAddr: "127.0.0.1:9094"}, Operation: "commit transaction", Err: errors.New("not primary")}
	if !strings.Contains(err.Error(), "127.0.0.1:9094") || !strings.Contains(err.Error(), "commit transaction") {
		t.Fatalf("error=%s", err.Error())
	}
}
