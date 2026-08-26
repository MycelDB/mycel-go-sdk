package mycel

import (
	"context"
	"errors"
	"fmt"
	"net"
	"syscall"
	"testing"

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
