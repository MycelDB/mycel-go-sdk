package mycel

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("MYCELD_GRPC_ADDR", "127.0.0.1:9999")
	t.Setenv("MYCEL_USERNAME", "user")
	t.Setenv("MYCEL_PASSWORD", "pass")
	t.Setenv("MYCEL_ACCESS_TOKEN", "token")
	t.Setenv("MYCELD_TLS", "true")
	t.Setenv("MYCELD_TLS_INSECURE_SKIP_VERIFY", "yes")
	t.Setenv("MYCEL_CLIENT_NAME", "bench")

	cfg := ConfigFromEnv()
	if cfg.Addr != "127.0.0.1:9999" || cfg.Username != "user" || cfg.Password != "pass" || cfg.AccessToken != "token" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if !cfg.TLS || !cfg.TLSInsecureSkipVerify || cfg.ClientName != "bench" {
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
