package mycel

import (
	"context"
	"testing"
	"time"

	commonv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/common/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("MYCELD_GRPC_ADDR", "127.0.0.1:9999")
	t.Setenv("MYCEL_USERNAME", "user")
	t.Setenv("MYCEL_PASSWORD", "pass")
	t.Setenv("MYCEL_ACCESS_TOKEN", "token")
	t.Setenv("MYCEL_ACCESS_TOKEN_EXPIRE_TIME", "2026-07-03T12:00:00Z")
	t.Setenv("MYCEL_REFRESH_TOKEN", "refresh")
	t.Setenv("MYCEL_REFRESH_BEFORE", "10s")
	t.Setenv("MYCEL_CALL_TIMEOUT", "5s")
	t.Setenv("MYCELD_TLS", "true")
	t.Setenv("MYCELD_TLS_INSECURE_SKIP_VERIFY", "yes")
	t.Setenv("MYCEL_CLIENT_NAME", "bench")

	cfg := ConfigFromEnv()
	if cfg.Addr != "127.0.0.1:9999" || cfg.Username != "user" || cfg.Password != "pass" || cfg.AccessToken != "token" || cfg.RefreshToken != "refresh" || cfg.RefreshBefore.String() != "10s" || cfg.CallTimeout.String() != "5s" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if !cfg.TLS || !cfg.TLSInsecureSkipVerify || cfg.ClientName != "bench" || !cfg.AccessTokenExpireTime.Equal(time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected TLS/client config: %+v", cfg)
	}
}

func TestTokenSourceContext(t *testing.T) {
	tokens := newTokenSource("abc")
	ctx := tokens.Context(context.Background())
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}
	got := md.Get(authorizationHeader)
	if len(got) != 1 || got[0] != "Bearer abc" {
		t.Fatalf("unexpected auth metadata: %+v", got)
	}
	tokens.Set("")
	ctx = tokens.Context(context.Background())
	md, _ = metadata.FromOutgoingContext(ctx)
	if len(md.Get(authorizationHeader)) != 0 {
		t.Fatalf("expected no auth metadata, got %+v", md.Get(authorizationHeader))
	}
}

func TestTransportOptionRejectsPartialClientCertificate(t *testing.T) {
	if _, err := transportOption(Config{TLS: true, TLSClientCertFile: "client.pem"}); err == nil {
		t.Fatal("expected partial client cert validation error")
	}
}

func TestClientAuthContext(t *testing.T) {
	c := &Client{tokens: newTokenSource("client-token")}
	ctx := c.AuthContext(context.Background())
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok || len(md.Get(authorizationHeader)) != 1 || md.Get(authorizationHeader)[0] != "Bearer client-token" {
		t.Fatalf("unexpected auth metadata: %+v", md)
	}
}

func TestAuthCallContextAddsTokenAndDeadline(t *testing.T) {
	c := &Client{tokens: newTokenSource("tok"), cfg: Config{CallTimeout: time.Second}}
	ctx, cancel := c.AuthCallContext(context.Background())
	defer cancel()
	if _, ok := ctx.Deadline(); !ok {
		t.Fatal("expected deadline")
	}
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok || len(md.Get(authorizationHeader)) != 1 || md.Get(authorizationHeader)[0] != "Bearer tok" {
		t.Fatalf("unexpected metadata: %+v", md)
	}
}

func TestTokenSourceRetriesExpiredUnauthenticatedOnce(t *testing.T) {
	tokens := newTokenSource("old-token")
	tokens.SetRefreshToken("refresh-token")
	refreshes := 0
	tokens.SetRefresher(func(context.Context) error {
		refreshes++
		tokens.SetAccessToken("new-token", time.Now().Add(time.Hour))
		return nil
	})

	calls := 0
	err := tokens.unaryInterceptor(context.Background(), "/svc/Protected", nil, nil, nil, func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		calls++
		md, _ := metadata.FromOutgoingContext(ctx)
		got := md.Get(authorizationHeader)
		if calls == 1 {
			if len(got) != 1 || got[0] != "Bearer old-token" {
				t.Fatalf("first call metadata = %+v", got)
			}
			return status.Error(codes.Unauthenticated, "authorization token is expired")
		}
		if len(got) != 1 || got[0] != "Bearer new-token" {
			t.Fatalf("retry metadata = %+v", got)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unaryInterceptor() error = %v", err)
	}
	if calls != 2 || refreshes != 1 {
		t.Fatalf("calls=%d refreshes=%d, want 2/1", calls, refreshes)
	}
}

func TestTokenSourceRetriesExpiredUnauthenticatedStreamOnce(t *testing.T) {
	tokens := newTokenSource("old-token")
	tokens.SetRefreshToken("refresh-token")
	refreshes := 0
	tokens.SetRefresher(func(context.Context) error {
		refreshes++
		tokens.SetAccessToken("new-token", time.Now().Add(time.Hour))
		return nil
	})

	calls := 0
	_, err := tokens.streamInterceptor(context.Background(), &grpc.StreamDesc{}, nil, "/svc/Stream", func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		calls++
		md, _ := metadata.FromOutgoingContext(ctx)
		got := md.Get(authorizationHeader)
		if calls == 1 {
			if len(got) != 1 || got[0] != "Bearer old-token" {
				t.Fatalf("first stream metadata = %+v", got)
			}
			return nil, status.Error(codes.Unauthenticated, "authorization token is expired")
		}
		if len(got) != 1 || got[0] != "Bearer new-token" {
			t.Fatalf("retry stream metadata = %+v", got)
		}
		return nil, nil
	})
	if err != nil {
		t.Fatalf("streamInterceptor() error = %v", err)
	}
	if calls != 2 || refreshes != 1 {
		t.Fatalf("calls=%d refreshes=%d, want 2/1", calls, refreshes)
	}
}

func TestTokenSourceContextReplacesExistingAuthorization(t *testing.T) {
	tokens := newTokenSource("new-token")
	ctx := metadata.AppendToOutgoingContext(context.Background(), authorizationHeader, "Bearer old-token")
	ctx = tokens.Context(ctx)
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}
	got := md.Get(authorizationHeader)
	if len(got) != 1 || got[0] != "Bearer new-token" {
		t.Fatalf("expected replacement authorization metadata, got %+v", got)
	}
}

func TestTokenSourceProactivelyRefreshesNearExpiry(t *testing.T) {
	tokens := newTokenSource("old-token")
	tokens.SetTokens("old-token", time.Now().Add(5*time.Second), "refresh-token")
	tokens.SetRefresher(func(context.Context) error {
		tokens.SetAccessToken("new-token", time.Now().Add(time.Hour))
		return nil
	})

	err := tokens.unaryInterceptor(context.Background(), "/svc/Protected", nil, nil, nil, func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		md, _ := metadata.FromOutgoingContext(ctx)
		got := md.Get(authorizationHeader)
		if len(got) != 1 || got[0] != "Bearer new-token" {
			t.Fatalf("metadata after proactive refresh = %+v", got)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unaryInterceptor() error = %v", err)
	}
}

func TestTokenSourceDoesNotRefreshAuthMethods(t *testing.T) {
	tokens := newTokenSource("old-token")
	tokens.SetTokens("old-token", time.Now().Add(-time.Second), "refresh-token")
	tokens.SetRefresher(func(context.Context) error {
		t.Fatal("auth method should not auto-refresh")
		return nil
	})
	err := tokens.unaryInterceptor(context.Background(), "/mycel.common.v1.AuthService/Refresh", nil, nil, nil, func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return status.Error(codes.Unauthenticated, "authorization token is expired")
	})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("error code = %s, want Unauthenticated", status.Code(err))
	}
}

func TestClientRefreshUsesStoredRefreshTokenAndRotatesState(t *testing.T) {
	expire := time.Now().Add(time.Hour).UTC()
	auth := &fakeClientAuth{refreshResponse: &commonv1.RefreshResponse{AccessToken: "access-2", AccessTokenExpireTime: timestamppb.New(expire), RefreshToken: ptr("refresh-2")}}
	c := &Client{Auth: auth, tokens: newTokenSource("access-1")}
	c.SetRefreshToken("refresh-1")
	res, err := c.Refresh(context.Background(), "")
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if res.GetAccessToken() != "access-2" || c.AccessToken() != "access-2" || c.RefreshToken() != "refresh-2" || !c.AccessTokenExpireTime().Equal(expire) {
		t.Fatalf("unexpected refreshed state token=%q refresh=%q expire=%s res=%#v", c.AccessToken(), c.RefreshToken(), c.AccessTokenExpireTime(), res)
	}
	if auth.refreshRequest == nil || auth.refreshRequest.GetRefreshToken() != "refresh-1" {
		t.Fatalf("refresh request did not use stored token: %#v", auth.refreshRequest)
	}
}

func TestAdminRefreshUsesStoredRefreshTokenAndRotatesState(t *testing.T) {
	expire := time.Now().Add(time.Hour).UTC()
	auth := &fakeAdminAuth{refreshResponse: &commonv1.RefreshResponse{AccessToken: "admin-access-2", AccessTokenExpireTime: timestamppb.New(expire), RefreshToken: ptr("admin-refresh-2")}}
	c := &AdminClient{Auth: auth, tokens: newTokenSource("admin-access-1")}
	c.SetRefreshToken("admin-refresh-1")
	res, err := c.Refresh(context.Background(), "")
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if res.GetAccessToken() != "admin-access-2" || c.AccessToken() != "admin-access-2" || c.RefreshToken() != "admin-refresh-2" || !c.AccessTokenExpireTime().Equal(expire) {
		t.Fatalf("unexpected refreshed admin state token=%q refresh=%q expire=%s res=%#v", c.AccessToken(), c.RefreshToken(), c.AccessTokenExpireTime(), res)
	}
	if auth.refreshRequest == nil || auth.refreshRequest.GetRefreshToken() != "admin-refresh-1" {
		t.Fatalf("refresh request did not use stored admin token: %#v", auth.refreshRequest)
	}
}

func TestAdminLogoutClearsCurrentAuthState(t *testing.T) {
	auth := &fakeAdminAuth{logoutResponse: &commonv1.LogoutResponse{}}
	c := &AdminClient{Auth: auth, tokens: newTokenSource("admin-access")}
	c.SetRefreshToken("admin-refresh")
	if _, err := c.Logout(context.Background(), ""); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if c.AccessToken() != "" || c.RefreshToken() != "" {
		t.Fatalf("expected auth state cleared, got access=%q refresh=%q", c.AccessToken(), c.RefreshToken())
	}
}

func ptr(value string) *string { return &value }

type fakeClientAuth struct {
	loginResponse   *commonv1.LoginResponse
	refreshResponse *commonv1.RefreshResponse
	logoutResponse  *commonv1.LogoutResponse
	refreshRequest  *commonv1.RefreshRequest
}

func (f *fakeClientAuth) Login(context.Context, *commonv1.LoginRequest, ...grpc.CallOption) (*commonv1.LoginResponse, error) {
	if f.loginResponse != nil {
		return f.loginResponse, nil
	}
	return &commonv1.LoginResponse{}, nil
}

func (f *fakeClientAuth) Refresh(_ context.Context, req *commonv1.RefreshRequest, _ ...grpc.CallOption) (*commonv1.RefreshResponse, error) {
	f.refreshRequest = req
	if f.refreshResponse != nil {
		return f.refreshResponse, nil
	}
	return &commonv1.RefreshResponse{}, nil
}

func (f *fakeClientAuth) Logout(context.Context, *commonv1.LogoutRequest, ...grpc.CallOption) (*commonv1.LogoutResponse, error) {
	if f.logoutResponse != nil {
		return f.logoutResponse, nil
	}
	return &commonv1.LogoutResponse{}, nil
}

func (f *fakeClientAuth) WhoAmI(context.Context, *commonv1.WhoAmIRequest, ...grpc.CallOption) (*commonv1.WhoAmIResponse, error) {
	return &commonv1.WhoAmIResponse{}, nil
}

func (f *fakeClientAuth) GetMyAccess(context.Context, *commonv1.GetMyAccessRequest, ...grpc.CallOption) (*commonv1.GetMyAccessResponse, error) {
	return &commonv1.GetMyAccessResponse{}, nil
}

func (f *fakeClientAuth) ListAuthSessions(context.Context, *commonv1.ListAuthSessionsRequest, ...grpc.CallOption) (*commonv1.ListAuthSessionsResponse, error) {
	return &commonv1.ListAuthSessionsResponse{}, nil
}

func (f *fakeClientAuth) RevokeAuthSession(context.Context, *commonv1.RevokeAuthSessionRequest, ...grpc.CallOption) (*commonv1.RevokeAuthSessionResponse, error) {
	return &commonv1.RevokeAuthSessionResponse{}, nil
}

func (f *fakeClientAuth) RevokeOtherAuthSessions(context.Context, *commonv1.RevokeOtherAuthSessionsRequest, ...grpc.CallOption) (*commonv1.RevokeOtherAuthSessionsResponse, error) {
	return &commonv1.RevokeOtherAuthSessionsResponse{}, nil
}

type fakeAdminAuth = fakeClientAuth
