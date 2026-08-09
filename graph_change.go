package mycel

import (
	"context"

	clientv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/client/v1"
	"google.golang.org/grpc"
)

// WatchGraphChanges opens the graph-change stream for one space/domain.
//
// The returned stream may deliver checkpoints, committed graph-change events,
// explicit gaps, and heartbeats. Operation IDs in graph-change origins are
// correlation metadata only.
func (c *Client) WatchGraphChanges(ctx context.Context, req *clientv1.WatchGraphChangesRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[clientv1.WatchGraphChangesResponse], error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	stream, err := c.GraphChange.WatchGraphChanges(callCtx, req, opts...)
	if err != nil {
		cancel()
		return nil, err
	}
	return &cancelOnCloseGraphChangeStream{ServerStreamingClient: stream, cancel: cancel}, nil
}

type cancelOnCloseGraphChangeStream struct {
	grpc.ServerStreamingClient[clientv1.WatchGraphChangesResponse]
	cancel context.CancelFunc
}

func (s *cancelOnCloseGraphChangeStream) Recv() (*clientv1.WatchGraphChangesResponse, error) {
	msg, err := s.ServerStreamingClient.Recv()
	if err != nil {
		s.cancel()
	}
	return msg, err
}
