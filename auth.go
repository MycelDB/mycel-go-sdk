package mycel

import (
	"context"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const authorizationHeader = "authorization"

const defaultRefreshBefore = 30 * time.Second

func skipAutoRefresh(method string) bool {
	switch method {
	case "/mycel.common.v1.AuthService/Login",
		"/mycel.common.v1.AuthService/Refresh",
		"/mycel.common.v1.AuthService/Logout":
		return true
	default:
		return false
	}
}

type tokenSource struct {
	mu                    sync.RWMutex
	refreshMu             sync.Mutex
	accessToken           string
	accessTokenExpireTime time.Time
	refreshToken          string
	refreshBefore         time.Duration
	refresher             func(context.Context) error
}

func newTokenSource(token string) *tokenSource {
	return &tokenSource{accessToken: strings.TrimSpace(token), refreshBefore: defaultRefreshBefore}
}

func newTokenSourceFromConfig(cfg Config) *tokenSource {
	s := newTokenSource(cfg.AccessToken)
	s.SetAccessToken(cfg.AccessToken, cfg.AccessTokenExpireTime)
	s.SetRefreshToken(cfg.RefreshToken)
	s.SetRefreshBefore(cfg.RefreshBefore)
	return s
}

func (s *tokenSource) Set(token string) {
	s.SetAccessToken(token, time.Time{})
}

func (s *tokenSource) SetAccessToken(token string, expireTime time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accessToken = strings.TrimSpace(token)
	s.accessTokenExpireTime = normalizeTime(expireTime)
}

func (s *tokenSource) SetRefreshToken(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshToken = strings.TrimSpace(token)
}

func (s *tokenSource) SetTokens(accessToken string, expireTime time.Time, refreshToken string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accessToken = strings.TrimSpace(accessToken)
	s.accessTokenExpireTime = normalizeTime(expireTime)
	if strings.TrimSpace(refreshToken) != "" {
		s.refreshToken = strings.TrimSpace(refreshToken)
	}
}

func (s *tokenSource) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accessToken = ""
	s.accessTokenExpireTime = time.Time{}
	s.refreshToken = ""
}

func (s *tokenSource) SetRefresher(refresher func(context.Context) error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refresher = refresher
}

func (s *tokenSource) SetRefreshBefore(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if d <= 0 {
		s.refreshBefore = defaultRefreshBefore
		return
	}
	s.refreshBefore = d
}

func (s *tokenSource) Token() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.accessToken
}

func (s *tokenSource) RefreshToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.refreshToken
}

func (s *tokenSource) AccessTokenExpireTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.accessTokenExpireTime
}

func (s *tokenSource) Context(ctx context.Context) context.Context {
	token := s.Token()
	md, _ := metadata.FromOutgoingContext(ctx)
	md = md.Copy()
	md.Delete(authorizationHeader)
	if token != "" {
		md.Set(authorizationHeader, "Bearer "+token)
	}
	return metadata.NewOutgoingContext(ctx, md)
}

func (s *tokenSource) unaryInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	if !skipAutoRefresh(method) {
		if err := s.refreshIfNeeded(ctx); err != nil && s.isExpired(time.Now()) {
			return err
		}
	}
	err := invoker(s.Context(ctx), method, req, reply, cc, opts...)
	if err == nil || skipAutoRefresh(method) || !isExpiredUnauthenticated(err) || !s.canRefresh() {
		return err
	}
	if refreshErr := s.refreshNow(ctx); refreshErr != nil {
		return refreshErr
	}
	return invoker(s.Context(ctx), method, req, reply, cc, opts...)
}

func (s *tokenSource) streamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	if !skipAutoRefresh(method) {
		if err := s.refreshIfNeeded(ctx); err != nil && s.isExpired(time.Now()) {
			return nil, err
		}
	}
	stream, err := streamer(s.Context(ctx), desc, cc, method, opts...)
	if err == nil || skipAutoRefresh(method) || !isExpiredUnauthenticated(err) || !s.canRefresh() {
		return stream, err
	}
	if refreshErr := s.refreshNow(ctx); refreshErr != nil {
		return nil, refreshErr
	}
	return streamer(s.Context(ctx), desc, cc, method, opts...)
}

func (s *tokenSource) refreshIfNeeded(ctx context.Context) error {
	if !s.needsRefresh(time.Now()) {
		return nil
	}
	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()
	if !s.needsRefresh(time.Now()) {
		return nil
	}
	return s.callRefresher(ctx)
}

func (s *tokenSource) refreshNow(ctx context.Context) error {
	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()
	return s.callRefresher(ctx)
}

func (s *tokenSource) callRefresher(ctx context.Context) error {
	refresher := s.currentRefresher()
	if refresher == nil || !s.canRefresh() {
		return nil
	}
	return refresher(ctx)
}

func (s *tokenSource) currentRefresher() func(context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.refresher
}

func (s *tokenSource) canRefresh() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.refresher != nil && s.refreshToken != ""
}

func (s *tokenSource) needsRefresh(now time.Time) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.refresher == nil || s.refreshToken == "" || s.accessTokenExpireTime.IsZero() {
		return false
	}
	refreshBefore := s.refreshBefore
	if refreshBefore <= 0 {
		refreshBefore = defaultRefreshBefore
	}
	return !now.Before(s.accessTokenExpireTime.Add(-refreshBefore))
}

func (s *tokenSource) isExpired(now time.Time) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return !s.accessTokenExpireTime.IsZero() && !now.Before(s.accessTokenExpireTime)
}

func isExpiredUnauthenticated(err error) bool {
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Unauthenticated {
		return false
	}
	return strings.Contains(strings.ToLower(st.Message()), "expired")
}

func normalizeTime(t time.Time) time.Time {
	if t.IsZero() {
		return time.Time{}
	}
	return t.UTC()
}
