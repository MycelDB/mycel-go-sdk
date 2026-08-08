package mycel

import (
	"context"
	"io"
	"testing"
	"time"

	clientv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/client/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestWatchGraphChangesForwardsRequestAndCancelsOnTerminalRecv(t *testing.T) {
	fake := &fakeGraphChangeServiceClient{stream: &fakeGraphChangeStream{err: io.EOF}}
	c := &Client{GraphChange: fake}
	req := &clientv1.WatchGraphChangesRequest{SpaceId: "space-1", DomainId: "domain-1", IncludeCurrent: true}

	stream, err := c.WatchGraphChanges(context.Background(), req)
	if err != nil {
		t.Fatalf("WatchGraphChanges() error = %v", err)
	}
	if fake.req != req {
		t.Fatalf("request was not forwarded")
	}
	if _, err := stream.Recv(); err != io.EOF {
		t.Fatalf("Recv() error = %v, want EOF", err)
	}
	select {
	case <-fake.ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("watch context was not canceled after terminal Recv")
	}
}

type fakeGraphChangeServiceClient struct {
	req    *clientv1.WatchGraphChangesRequest
	ctx    context.Context
	stream grpc.ServerStreamingClient[clientv1.WatchGraphChangesResponse]
	err    error
}

func (f *fakeGraphChangeServiceClient) WatchGraphChanges(ctx context.Context, in *clientv1.WatchGraphChangesRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[clientv1.WatchGraphChangesResponse], error) {
	f.ctx = ctx
	f.req = in
	if f.err != nil {
		return nil, f.err
	}
	return f.stream, nil
}

type fakeGraphChangeStream struct {
	msg *clientv1.WatchGraphChangesResponse
	err error
}

func (s *fakeGraphChangeStream) Recv() (*clientv1.WatchGraphChangesResponse, error) {
	return s.msg, s.err
}

func (s *fakeGraphChangeStream) Header() (metadata.MD, error) { return nil, nil }
func (s *fakeGraphChangeStream) Trailer() metadata.MD         { return nil }
func (s *fakeGraphChangeStream) CloseSend() error             { return nil }
func (s *fakeGraphChangeStream) Context() context.Context     { return context.Background() }
func (s *fakeGraphChangeStream) SendMsg(m any) error          { return nil }
func (s *fakeGraphChangeStream) RecvMsg(m any) error          { return nil }
