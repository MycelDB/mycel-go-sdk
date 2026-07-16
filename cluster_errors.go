package mycel

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	NotPrimaryReason   = "MYCEL_CLUSTER_NOT_PRIMARY"
	PrimaryNodeIDKey   = "mycel-primary-node-id"
	PrimaryNodeNameKey = "mycel-primary-node-name"
	PrimaryBackendKey  = "mycel-primary-backend-advertise-addr"
	AuthorityEpochKey  = "mycel-authority-epoch"
)

type PrimaryHint struct {
	NodeID               string
	NodeName             string
	BackendAdvertiseAddr string
	AuthorityEpoch       string
}

func IsNotPrimaryError(err error) bool {
	st, ok := status.FromError(err)
	return ok && st.Code() == codes.FailedPrecondition && st.Message() == "node is not cluster primary"
}

func PrimaryHintFromError(err error) (PrimaryHint, bool) {
	st, ok := status.FromError(err)
	if !ok {
		return PrimaryHint{}, false
	}
	for _, detail := range st.Details() {
		info, ok := detail.(*errdetails.ErrorInfo)
		if !ok || info.GetReason() != NotPrimaryReason {
			continue
		}
		md := info.GetMetadata()
		return PrimaryHint{NodeID: md[PrimaryNodeIDKey], NodeName: md[PrimaryNodeNameKey], BackendAdvertiseAddr: md[PrimaryBackendKey], AuthorityEpoch: md[AuthorityEpochKey]}, true
	}
	return PrimaryHint{}, false
}
