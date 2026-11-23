package mycel

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"strings"
	"syscall"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorKind is a stable SDK error category for downstream applications.
type ErrorKind string

const (
	// AuthErrorInfoDomain is the google.rpc.ErrorInfo domain used by Mycel daemon
	// authentication and auth-session errors.
	AuthErrorInfoDomain = "mycel.auth"

	// AuthReasonAuthSessionInvalid means the access token refers to no usable durable auth session.
	AuthReasonAuthSessionInvalid = "auth_session_invalid"
	// AuthReasonAuthSessionRevoked means the durable auth session or token family was revoked.
	AuthReasonAuthSessionRevoked = "auth_session_revoked"
	// AuthReasonAuthSessionExpired means the durable auth session exceeded idle or absolute expiry.
	AuthReasonAuthSessionExpired = "auth_session_expired"
	// AuthReasonRefreshSessionInvalid means the presented refresh token/session is unusable.
	AuthReasonRefreshSessionInvalid = "refresh_session_invalid"

	// BackupErrorInfoDomain is the google.rpc.ErrorInfo domain used by Mycel daemon
	// backup, restore, and quiesce errors.
	BackupErrorInfoDomain = "mycel.backup"

	// BackupReasonClusterQuiesced means the daemon is intentionally rejecting work
	// while backup quiescence is active.
	BackupReasonClusterQuiesced = "cluster_quiesced"
	// BackupReasonTransactionAbortedByQuiesce means a transaction finalization was
	// aborted because backup quiescence closed admission first.
	BackupReasonTransactionAbortedByQuiesce = "transaction_aborted_by_quiesce"

	// UsageErrorInfoDomain is the google.rpc.ErrorInfo domain used by Mycel usage
	// hard-limit admission errors.
	UsageErrorInfoDomain = "mycel.usage"
	// UsageReasonLimitExceeded means a write would exceed a configured hard usage limit.
	UsageReasonLimitExceeded = "usage_limit_exceeded"
	// UsageReasonWriteGateBlocked means a configured usage write gate is fail-closed
	// because quota-safe accounting is not currently ready.
	UsageReasonWriteGateBlocked = "usage_write_gate_blocked"
)

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

// IsAuthSessionInvalid reports whether err is an Unauthenticated gRPC status
// that identifies a stale, revoked, expired, or otherwise unusable durable auth
// session. Long-running services that cache delegated clients should discard the
// cached client/session when this returns true, then create a new delegated
// session if application policy still permits acting for that principal.
func IsAuthSessionInvalid(err error) bool {
	_, ok := AuthSessionInvalidReason(err)
	return ok
}

// AuthSessionInvalidReason returns the stable Mycel auth-session ErrorInfo
// reason when err indicates a cached auth/delegated session should be discarded.
func AuthSessionInvalidReason(err error) (string, bool) {
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Unauthenticated {
		return "", false
	}
	for _, detail := range st.Details() {
		info, ok := detail.(*errdetails.ErrorInfo)
		if !ok || info.GetDomain() != AuthErrorInfoDomain {
			continue
		}
		reason := strings.ToLower(strings.TrimSpace(info.GetReason()))
		if isAuthSessionInvalidReason(reason) {
			return reason, true
		}
	}
	return "", false
}

func isAuthSessionInvalidReason(reason string) bool {
	switch reason {
	case AuthReasonAuthSessionInvalid, AuthReasonAuthSessionRevoked, AuthReasonAuthSessionExpired, AuthReasonRefreshSessionInvalid:
		return true
	default:
		return false
	}
}

// IsClusterQuiesced reports whether err is an Unavailable gRPC status carrying
// the stable Mycel backup/quiesce cluster_quiesced reason. It intentionally does
// not classify PermissionDenied, NotFound, or generic Unavailable errors.
func IsClusterQuiesced(err error) bool {
	reason, ok := BackupQuiesceReason(err)
	return ok && reason == BackupReasonClusterQuiesced
}

// IsTransactionAbortedByQuiesce reports whether err indicates a write
// transaction finalization was aborted by backup quiescence.
func IsTransactionAbortedByQuiesce(err error) bool {
	reason, ok := BackupQuiesceReason(err)
	return ok && reason == BackupReasonTransactionAbortedByQuiesce
}

// BackupQuiesceReason returns a stable Mycel backup/quiesce ErrorInfo reason
// from an Unavailable gRPC status when present.
func BackupQuiesceReason(err error) (string, bool) {
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Unavailable {
		return "", false
	}
	for _, detail := range st.Details() {
		info, ok := detail.(*errdetails.ErrorInfo)
		if !ok || info.GetDomain() != BackupErrorInfoDomain {
			continue
		}
		reason := strings.ToLower(strings.TrimSpace(info.GetReason()))
		switch reason {
		case BackupReasonClusterQuiesced, BackupReasonTransactionAbortedByQuiesce:
			return reason, true
		}
	}
	return "", false
}

// IsUsageLimitExceeded reports whether err is a ResourceExhausted gRPC status
// carrying the stable Mycel usage_limit_exceeded reason.
func IsUsageLimitExceeded(err error) bool {
	reason, ok := UsageAdmissionReason(err)
	return ok && reason == UsageReasonLimitExceeded
}

// IsUsageWriteGateBlocked reports whether err is a ResourceExhausted gRPC
// status carrying the stable Mycel usage_write_gate_blocked reason.
func IsUsageWriteGateBlocked(err error) bool {
	reason, ok := UsageAdmissionReason(err)
	return ok && reason == UsageReasonWriteGateBlocked
}

// UsageAdmissionReason returns a stable Mycel usage admission ErrorInfo reason
// from a ResourceExhausted gRPC status when present.
func UsageAdmissionReason(err error) (string, bool) {
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.ResourceExhausted {
		return "", false
	}
	for _, detail := range st.Details() {
		info, ok := detail.(*errdetails.ErrorInfo)
		if !ok || info.GetDomain() != UsageErrorInfoDomain {
			continue
		}
		reason := strings.ToLower(strings.TrimSpace(info.GetReason()))
		switch reason {
		case UsageReasonLimitExceeded, UsageReasonWriteGateBlocked:
			return reason, true
		}
	}
	return "", false
}

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
