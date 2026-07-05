package mycel

import (
	"context"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const authorizationHeader = "authorization"

type tokenSource struct {
	mu    sync.RWMutex
	token string
}

func newTokenSource(token string) *tokenSource {
	return &tokenSource{token: strings.TrimSpace(token)}
}

func (s *tokenSource) Set(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.token = strings.TrimSpace(token)
}

func (s *tokenSource) Token() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.token
}

func (s *tokenSource) Context(ctx context.Context) context.Context {
	if token := s.Token(); token != "" {
		return metadata.AppendToOutgoingContext(ctx, authorizationHeader, "Bearer "+token)
	}
	return ctx
}

func (s *tokenSource) unaryInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return invoker(s.Context(ctx), method, req, reply, cc, opts...)
}

func (s *tokenSource) streamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return streamer(s.Context(ctx), desc, cc, method, opts...)
}
