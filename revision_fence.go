package mycel

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const revisionFenceConflictReason = "GRAPH_REVISION_FENCE_CONFLICT"

// RevisionFenceConflictError describes a commit rejected because one or more
// read dependencies changed. The daemon applies no mutation in this case.
type RevisionFenceConflictError struct {
	Status *status.Status
}

func (e *RevisionFenceConflictError) Error() string {
	if e == nil || e.Status == nil {
		return "graph revision fence conflict"
	}
	return e.Status.Message()
}

// AsRevisionFenceConflict returns typed conflict information without replacing
// the original gRPC status error, so status.Code remains ABORTED.
func AsRevisionFenceConflict(err error) (*RevisionFenceConflictError, bool) {
	if err == nil || status.Code(err) != codes.Aborted {
		return nil, false
	}
	st := status.Convert(err)
	for _, detail := range st.Details() {
		info, ok := detail.(*errdetails.ErrorInfo)
		if ok && info.GetReason() == revisionFenceConflictReason {
			return &RevisionFenceConflictError{Status: st}, true
		}
	}
	return nil, false
}

// IsRevisionFenceConflict reports whether err is the daemon's typed revision
// fence conflict. Callers should re-read and regenerate derived output rather
// than retrying the stale mutation.
func IsRevisionFenceConflict(err error) bool {
	_, ok := AsRevisionFenceConflict(err)
	return ok
}

func cloneRevisionWitnesses(in [][]byte) [][]byte {
	if len(in) == 0 {
		return nil
	}
	out := make([][]byte, len(in))
	for i, witness := range in {
		if witness != nil {
			out[i] = append([]byte(nil), witness...)
		}
	}
	return out
}
