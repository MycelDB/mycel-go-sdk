package mycel

import (
	"context"
	"errors"
	"fmt"
	"net"
	"syscall"
	"testing"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type timeoutError struct{}

func (timeoutError) Error() string   { return "network timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

func TestClassifyErrorMapsGRPCStatusCodes(t *testing.T) {
	cases := []struct {
		code     codes.Code
		kind     ErrorKind
		severity ErrorSeverity
	}{
		{codes.InvalidArgument, ErrorKindValidation, ErrorSeverityWarning},
		{codes.Unauthenticated, ErrorKindAuthentication, ErrorSeverityWarning},
		{codes.PermissionDenied, ErrorKindAuthorization, ErrorSeverityWarning},
		{codes.NotFound, ErrorKindNotFound, ErrorSeverityWarning},
		{codes.AlreadyExists, ErrorKindConflict, ErrorSeverityWarning},
		{codes.Aborted, ErrorKindConflict, ErrorSeverityWarning},
		{codes.ResourceExhausted, ErrorKindRateLimited, ErrorSeverityWarning},
		{codes.Unavailable, ErrorKindUnavailable, ErrorSeverityError},
		{codes.DeadlineExceeded, ErrorKindTimeout, ErrorSeverityError},
		{codes.Internal, ErrorKindInternal, ErrorSeverityError},
		{codes.DataLoss, ErrorKindInternal, ErrorSeverityError},
		{codes.Canceled, ErrorKindUnknown, ErrorSeverityError},
	}
	for _, tc := range cases {
		t.Run(tc.code.String(), func(t *testing.T) {
			classified := ClassifyError(status.Error(tc.code, "classified message"))
			if classified.Kind != tc.kind || classified.Severity != tc.severity || classified.Message != "classified message" {
				t.Fatalf("ClassifyError(%s) = %+v, want kind=%s severity=%s message=classified message", tc.code, classified, tc.kind, tc.severity)
			}
		})
	}
}

func TestClassifyErrorHandlesWrappedGRPCStatus(t *testing.T) {
	err := fmt.Errorf("during login: %w", status.Error(codes.PermissionDenied, "denied"))
	classified := ClassifyError(err)
	if classified.Kind != ErrorKindAuthorization || classified.Severity != ErrorSeverityWarning {
		t.Fatalf("wrapped status classified as %+v, want authorization/warning", classified)
	}
	if !IsAuthorizationError(err) || IsAuthenticationError(err) || IsConnectivityError(err) {
		t.Fatalf("unexpected helper predicate results for %v", err)
	}

	authErr := fmt.Errorf("auth: %w", status.Error(codes.Unauthenticated, "invalid credentials"))
	if !IsAuthenticationError(authErr) {
		t.Fatalf("wrapped unauthenticated status was not authentication")
	}
}

func TestIsAuthSessionInvalidUsesStableErrorInfoReason(t *testing.T) {
	st := status.New(codes.Unauthenticated, "auth session is revoked")
	withDetails, err := st.WithDetails(&errdetails.ErrorInfo{Domain: AuthErrorInfoDomain, Reason: AuthReasonAuthSessionRevoked})
	if err != nil {
		t.Fatalf("WithDetails() error = %v", err)
	}
	wrapped := fmt.Errorf("cached delegated client failed: %w", withDetails.Err())
	if !IsAuthSessionInvalid(wrapped) {
		t.Fatalf("IsAuthSessionInvalid() = false for %v", wrapped)
	}
	if reason, ok := AuthSessionInvalidReason(wrapped); !ok || reason != AuthReasonAuthSessionRevoked {
		t.Fatalf("AuthSessionInvalidReason() = %q, %v; want %q, true", reason, ok, AuthReasonAuthSessionRevoked)
	}

	accessExpired, err := status.New(codes.Unauthenticated, "authorization token is expired").WithDetails(&errdetails.ErrorInfo{Domain: AuthErrorInfoDomain, Reason: "authorization_token_expired"})
	if err != nil {
		t.Fatalf("WithDetails() error = %v", err)
	}
	if IsAuthSessionInvalid(accessExpired.Err()) {
		t.Fatal("access-token expiry alone should not be classified as invalid auth session")
	}
	if IsAuthSessionInvalid(status.Error(codes.PermissionDenied, "principal management capability is required")) {
		t.Fatal("permission denied should not be classified as invalid auth session")
	}
}

func TestBackupQuiesceReasonUsesStableErrorInfoReason(t *testing.T) {
	clusterStatus, err := status.New(codes.Unavailable, "myceld is temporarily quiesced").WithDetails(&errdetails.ErrorInfo{Domain: BackupErrorInfoDomain, Reason: BackupReasonClusterQuiesced})
	if err != nil {
		t.Fatalf("WithDetails() error = %v", err)
	}
	if !IsClusterQuiesced(clusterStatus.Err()) {
		t.Fatal("IsClusterQuiesced() = false for cluster_quiesced")
	}
	if IsTransactionAbortedByQuiesce(clusterStatus.Err()) {
		t.Fatal("cluster_quiesced should not be transaction_aborted_by_quiesce")
	}
	if reason, ok := BackupQuiesceReason(clusterStatus.Err()); !ok || reason != BackupReasonClusterQuiesced {
		t.Fatalf("BackupQuiesceReason() = %q, %v", reason, ok)
	}

	txStatus, err := status.New(codes.Unavailable, "transaction aborted by backup quiescence").WithDetails(&errdetails.ErrorInfo{Domain: BackupErrorInfoDomain, Reason: BackupReasonTransactionAbortedByQuiesce})
	if err != nil {
		t.Fatalf("WithDetails() error = %v", err)
	}
	if !IsTransactionAbortedByQuiesce(txStatus.Err()) {
		t.Fatal("IsTransactionAbortedByQuiesce() = false for transaction_aborted_by_quiesce")
	}
	if IsClusterQuiesced(status.Error(codes.Unavailable, "temporarily unavailable")) {
		t.Fatal("generic Unavailable should not be classified as cluster_quiesced")
	}
	if IsClusterQuiesced(status.Error(codes.PermissionDenied, "backup manage required")) || IsClusterQuiesced(status.Error(codes.NotFound, "backup set missing")) {
		t.Fatal("PermissionDenied/NotFound should not be quiesce signals")
	}
}

func TestUsageAdmissionReasonUsesStableErrorInfoReason(t *testing.T) {
	limitStatus, err := status.New(codes.ResourceExhausted, "usage admission rejected").WithDetails(&errdetails.ErrorInfo{Domain: UsageErrorInfoDomain, Reason: UsageReasonLimitExceeded})
	if err != nil {
		t.Fatalf("WithDetails() error = %v", err)
	}
	if !IsUsageLimitExceeded(limitStatus.Err()) {
		t.Fatal("IsUsageLimitExceeded() = false for usage_limit_exceeded")
	}
	if IsUsageWriteGateBlocked(limitStatus.Err()) {
		t.Fatal("usage_limit_exceeded should not be usage_write_gate_blocked")
	}
	if reason, ok := UsageAdmissionReason(limitStatus.Err()); !ok || reason != UsageReasonLimitExceeded {
		t.Fatalf("UsageAdmissionReason() = %q, %v", reason, ok)
	}

	gateStatus, err := status.New(codes.ResourceExhausted, "usage write gate blocked").WithDetails(&errdetails.ErrorInfo{Domain: UsageErrorInfoDomain, Reason: UsageReasonWriteGateBlocked})
	if err != nil {
		t.Fatalf("WithDetails() error = %v", err)
	}
	if !IsUsageWriteGateBlocked(gateStatus.Err()) {
		t.Fatal("IsUsageWriteGateBlocked() = false for usage_write_gate_blocked")
	}
	if IsUsageLimitExceeded(status.Error(codes.ResourceExhausted, "generic capacity")) || IsUsageLimitExceeded(status.Error(codes.PermissionDenied, "usage limit manage required")) {
		t.Fatal("generic ResourceExhausted/PermissionDenied should not be usage admission signals")
	}
}

func TestClassifyErrorHandlesTimeouts(t *testing.T) {
	for _, err := range []error{context.DeadlineExceeded, fmt.Errorf("call failed: %w", context.DeadlineExceeded), timeoutError{}} {
		classified := ClassifyError(err)
		if classified.Kind != ErrorKindTimeout || classified.Severity != ErrorSeverityError {
			t.Fatalf("ClassifyError(%T) = %+v, want timeout/error", err, classified)
		}
	}
}

func TestClassifyErrorHandlesConnectivity(t *testing.T) {
	cases := []error{
		&net.DNSError{Err: "no such host", Name: "bad.example"},
		&net.OpError{Op: "dial", Net: "tcp", Err: syscall.ECONNREFUSED},
		fmt.Errorf("dial: %w", syscall.ECONNREFUSED),
	}
	for _, err := range cases {
		classified := ClassifyError(err)
		if classified.Kind != ErrorKindConnectivity || classified.Severity != ErrorSeverityError {
			t.Fatalf("ClassifyError(%T) = %+v, want connectivity/error", err, classified)
		}
		if !IsConnectivityError(err) {
			t.Fatalf("IsConnectivityError(%T) = false", err)
		}
	}
}

func TestClassifyErrorUnknownFallbackAndNil(t *testing.T) {
	for _, err := range []error{errors.New("opaque"), nil} {
		classified := ClassifyError(err)
		if classified.Kind != ErrorKindUnknown || classified.Severity != ErrorSeverityError || classified.Message == "" {
			t.Fatalf("ClassifyError(%v) = %+v, want unknown/error with message", err, classified)
		}
	}
}

func TestDefaultSeverity(t *testing.T) {
	if DefaultSeverity(ErrorKindAuthentication) != ErrorSeverityWarning {
		t.Fatalf("authentication severity = %s, want warning", DefaultSeverity(ErrorKindAuthentication))
	}
	if DefaultSeverity(ErrorKindConnectivity) != ErrorSeverityError {
		t.Fatalf("connectivity severity = %s, want error", DefaultSeverity(ErrorKindConnectivity))
	}
}
