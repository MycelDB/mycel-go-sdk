package mycel

import (
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	NotPrimaryReason       = "MYCEL_CLUSTER_NOT_PRIMARY"
	SnapshotRequiredReason = "MYCEL_WAL_SNAPSHOT_REQUIRED"
	PrimaryNodeIDKey       = "mycel-primary-node-id"
	PrimaryNodeNameKey     = "mycel-primary-node-name"
	PrimaryBackendKey      = "mycel-primary-backend-advertise-addr"
	AuthorityEpochKey      = "mycel-authority-epoch"
)

type SnapshotRequiredInfo struct {
	RequestedAfterLSN uint64
	NextRequestedLSN  uint64
	FirstRetainedLSN  uint64
	LastCommittedLSN  uint64
	CheckpointLSN     uint64
	PrimaryNodeID     string
	AuthorityEpoch    string
}

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

func IsSnapshotRequiredError(err error) bool {
	st, ok := status.FromError(err)
	return ok && st.Code() == codes.FailedPrecondition && st.Message() == "follower requires snapshot catch-up"
}

func SnapshotRequiredInfoFromError(err error) (SnapshotRequiredInfo, bool) {
	st, ok := status.FromError(err)
	if !ok {
		return SnapshotRequiredInfo{}, false
	}
	for _, detail := range st.Details() {
		info, ok := detail.(*errdetails.ErrorInfo)
		if !ok || info.GetReason() != SnapshotRequiredReason {
			continue
		}
		md := info.GetMetadata()
		parse := func(k string) uint64 { var v uint64; fmt.Sscanf(md[k], "%d", &v); return v }
		return SnapshotRequiredInfo{RequestedAfterLSN: parse("requested_after_lsn"), NextRequestedLSN: parse("next_requested_lsn"), FirstRetainedLSN: parse("first_retained_lsn"), LastCommittedLSN: parse("last_committed_lsn"), CheckpointLSN: parse("checkpoint_lsn"), PrimaryNodeID: md["primary_node_id"], AuthorityEpoch: md["authority_epoch"]}, true
	}
	return SnapshotRequiredInfo{}, false
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
