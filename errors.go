package mycel

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"syscall"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorKind is a stable SDK error category for downstream applications.
type ErrorKind string

const (
	// ErrorKindValidation indicates invalid local input or invalid RPC arguments.
	ErrorKindValidation ErrorKind = "validation"
	// ErrorKindConnectivity indicates a local network, DNS, or TLS connectivity failure.
	ErrorKindConnectivity ErrorKind = "connectivity"
	// ErrorKindAuthentication indicates invalid, missing, or expired authentication.
	ErrorKindAuthentication ErrorKind = "authentication"
	// ErrorKindAuthorization indicates an authenticated principal lacks required access.
	ErrorKindAuthorization ErrorKind = "authorization"
	// ErrorKindNotFound indicates the requested resource does not exist.
	ErrorKindNotFound ErrorKind = "not_found"
	// ErrorKindConflict indicates a conflicting state or already-existing resource.
	ErrorKindConflict ErrorKind = "conflict"
	// ErrorKindRateLimited indicates a quota or rate limit was exceeded.
	ErrorKindRateLimited ErrorKind = "rate_limited"
	// ErrorKindUnavailable indicates the remote service reported unavailability.
	ErrorKindUnavailable ErrorKind = "unavailable"
	// ErrorKindTimeout indicates a deadline or timeout was reached.
	ErrorKindTimeout ErrorKind = "timeout"
	// ErrorKindInternal indicates an internal or data-loss error.
	ErrorKindInternal ErrorKind = "internal"
	// ErrorKindUnknown indicates the SDK could not identify a more specific category.
	ErrorKindUnknown ErrorKind = "unknown"
)

// ErrorSeverity is a default presentation severity hint for classified errors.
type ErrorSeverity string

const (
	// ErrorSeverityInfo is informational.
	ErrorSeverityInfo ErrorSeverity = "info"
	// ErrorSeverityWarning is user-correctable or expected operational feedback.
	ErrorSeverityWarning ErrorSeverity = "warning"
	// ErrorSeverityError is blocking or unexpected failure feedback.
	ErrorSeverityError ErrorSeverity = "error"
)

// ClassifiedError describes an error using stable application-facing fields.
type ClassifiedError struct {
	Kind     ErrorKind
	Severity ErrorSeverity
	Message  string
	Detail   string
}

// DefaultSeverity returns the default presentation severity for kind.
func DefaultSeverity(kind ErrorKind) ErrorSeverity {
	switch kind {
	case ErrorKindValidation, ErrorKindAuthentication, ErrorKindAuthorization, ErrorKindNotFound, ErrorKindConflict, ErrorKindRateLimited:
		return ErrorSeverityWarning
	case ErrorKindConnectivity, ErrorKindUnavailable, ErrorKindTimeout, ErrorKindInternal, ErrorKindUnknown:
		return ErrorSeverityError
	default:
		return ErrorSeverityError
	}
}

// ClassifyError classifies err into a stable SDK error category.
func ClassifyError(err error) ClassifiedError {
	if err == nil {
		return newClassifiedError(ErrorKindUnknown, "unknown error", "")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return newClassifiedError(ErrorKindTimeout, err.Error(), "")
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return newClassifiedError(ErrorKindTimeout, err.Error(), "")
	}
	if st, ok := status.FromError(err); ok {
		return classifyStatusError(err, st)
	}
	if isConnectivityError(err) {
		return newClassifiedError(ErrorKindConnectivity, err.Error(), "")
	}
	return newClassifiedError(ErrorKindUnknown, err.Error(), "")
}

// ErrorKindOf returns the stable error kind for err.
func ErrorKindOf(err error) ErrorKind { return ClassifyError(err).Kind }

// IsConnectivityError reports whether err is a local connectivity failure.
func IsConnectivityError(err error) bool { return ErrorKindOf(err) == ErrorKindConnectivity }

// IsAuthenticationError reports whether err is an authentication failure.
func IsAuthenticationError(err error) bool { return ErrorKindOf(err) == ErrorKindAuthentication }

// IsAuthorizationError reports whether err is an authorization failure.
func IsAuthorizationError(err error) bool { return ErrorKindOf(err) == ErrorKindAuthorization }

func classifyStatusError(err error, st *status.Status) ClassifiedError {
	kind := ErrorKindUnknown
	switch st.Code() {
	case codes.InvalidArgument:
		kind = ErrorKindValidation
	case codes.Unauthenticated:
		kind = ErrorKindAuthentication
	case codes.PermissionDenied:
		kind = ErrorKindAuthorization
	case codes.NotFound:
		kind = ErrorKindNotFound
	case codes.AlreadyExists, codes.Aborted:
		kind = ErrorKindConflict
	case codes.ResourceExhausted:
		kind = ErrorKindRateLimited
	case codes.Unavailable:
		kind = ErrorKindUnavailable
	case codes.DeadlineExceeded:
		kind = ErrorKindTimeout
	case codes.Internal, codes.DataLoss:
		kind = ErrorKindInternal
	}
	message := st.Message()
	if message == "" {
		message = st.Code().String()
	}
	detail := err.Error()
	if detail == message {
		detail = ""
	}
	return newClassifiedError(kind, message, detail)
}

func isConnectivityError(err error) bool {
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	if errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.EHOSTUNREACH) || errors.Is(err, syscall.ENETUNREACH) || errors.Is(err, syscall.EPIPE) {
		return true
	}
	var recordHeader tls.RecordHeaderError
	if errors.As(err, &recordHeader) {
		return true
	}
	var unknownAuthority x509.UnknownAuthorityError
	if errors.As(err, &unknownAuthority) {
		return true
	}
	var hostname x509.HostnameError
	if errors.As(err, &hostname) {
		return true
	}
	var certInvalid x509.CertificateInvalidError
	return errors.As(err, &certInvalid)
}

func newClassifiedError(kind ErrorKind, message, detail string) ClassifiedError {
	if message == "" {
		message = "unknown error"
	}
	return ClassifiedError{Kind: kind, Severity: DefaultSeverity(kind), Message: message, Detail: detail}
}
