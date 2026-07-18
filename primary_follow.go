package mycel

import (
	"context"
	"fmt"
	"strings"

	adminv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/admin/v1"
	clientv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/client/v1"
	"google.golang.org/grpc"
)

type PrimaryChangedRetryRequiredError struct {
	Hint      PrimaryHint
	Operation string
	Err       error
}

func (e *PrimaryChangedRetryRequiredError) Error() string {
	if e.Operation != "" {
		return fmt.Sprintf("primary changed to %s; retry %s on new primary: %v", e.Hint.BackendAdvertiseAddr, e.Operation, e.Err)
	}
	return fmt.Sprintf("primary changed to %s; retry operation on new primary: %v", e.Hint.BackendAdvertiseAddr, e.Err)
}
func (e *PrimaryChangedRetryRequiredError) Unwrap() error { return e.Err }

func (c *Client) CurrentAddr() string {
	if c == nil {
		return ""
	}
	return c.cfg.addr()
}
func (c *AdminClient) CurrentAddr() string {
	if c == nil {
		return ""
	}
	return c.cfg.addr()
}

func (c *Client) FollowPrimaryFromError(ctx context.Context, err error, opts ...grpc.DialOption) (PrimaryHint, bool, error) {
	hint, ok := PrimaryHintFromError(err)
	if !ok || strings.TrimSpace(hint.BackendAdvertiseAddr) == "" {
		return PrimaryHint{}, false, nil
	}
	return hint, true, c.Reconnect(ctx, hint.BackendAdvertiseAddr, opts...)
}
func (c *AdminClient) FollowPrimaryFromError(ctx context.Context, err error, opts ...grpc.DialOption) (PrimaryHint, bool, error) {
	hint, ok := PrimaryHintFromError(err)
	if !ok || strings.TrimSpace(hint.BackendAdvertiseAddr) == "" {
		return PrimaryHint{}, false, nil
	}
	return hint, true, c.Reconnect(ctx, hint.BackendAdvertiseAddr, opts...)
}

func (c *Client) Reconnect(ctx context.Context, addr string, opts ...grpc.DialOption) error {
	if c == nil {
		return fmt.Errorf("client is nil")
	}
	cfg := c.cfg
	cfg.Addr = strings.TrimSpace(addr)
	conn, err := dial(ctx, cfg, c.tokens, opts...)
	if err != nil {
		return err
	}
	old := c.Conn
	c.Conn = conn
	c.cfg = cfg
	c.Auth = clientv1.NewAuthServiceClient(conn)
	c.Space = clientv1.NewSpaceServiceClient(conn)
	c.Domain = clientv1.NewDomainServiceClient(conn)
	c.Template = clientv1.NewTemplateServiceClient(conn)
	c.Session = clientv1.NewSessionServiceClient(conn)
	c.Transaction = clientv1.NewTransactionServiceClient(conn)
	c.Graph = clientv1.NewGraphServiceClient(conn)
	c.Blob = clientv1.NewBlobServiceClient(conn)
	c.Query = clientv1.NewQueryServiceClient(conn)
	c.ImportExport = clientv1.NewImportExportServiceClient(conn)
	c.Metadata = clientv1.NewMetadataCatalogServiceClient(conn)
	c.Semantic = clientv1.NewSemanticServiceClient(conn)
	c.ChangeStream = clientv1.NewChangeStreamServiceClient(conn)
	if old != nil {
		_ = old.Close()
	}
	if cfg.Username != "" || cfg.Password != "" {
		_, err = c.Login(ctx, cfg.Username, cfg.Password)
	}
	return err
}

func (c *AdminClient) Reconnect(ctx context.Context, addr string, opts ...grpc.DialOption) error {
	if c == nil {
		return fmt.Errorf("admin client is nil")
	}
	cfg := c.cfg
	cfg.Addr = strings.TrimSpace(addr)
	conn, err := dial(ctx, cfg, c.tokens, opts...)
	if err != nil {
		return err
	}
	old := c.Conn
	c.Conn = conn
	c.cfg = cfg
	c.Auth = adminv1.NewAdminAuthServiceClient(conn)
	c.Operators = adminv1.NewAdminOperatorServiceClient(conn)
	c.Users = adminv1.NewAdminUserServiceClient(conn)
	c.Spaces = adminv1.NewAdminSpaceServiceClient(conn)
	c.Domains = adminv1.NewAdminDomainServiceClient(conn)
	c.Semantic = adminv1.NewAdminSemanticServiceClient(conn)
	c.SemanticMaintenance = adminv1.NewAdminSemanticMaintenanceServiceClient(conn)
	c.SemanticMigration = adminv1.NewAdminSemanticMigrationServiceClient(conn)
	c.Inference = adminv1.NewAdminInferenceServiceClient(conn)
	c.Backup = adminv1.NewAdminBackupServiceClient(conn)
	if old != nil {
		_ = old.Close()
	}
	if cfg.Username != "" || cfg.Password != "" {
		_, err = c.LoginOperator(ctx, cfg.Username, cfg.Password)
	}
	return err
}

func (c *Client) DoRead(ctx context.Context, operation string, fn func() error, opts ...grpc.DialOption) error {
	_, err := DoReadValue[struct{}](ctx, c, operation, func() (struct{}, error) { return struct{}{}, fn() }, opts...)
	return err
}

func DoReadValue[T any](ctx context.Context, c *Client, operation string, fn func() (T, error), opts ...grpc.DialOption) (T, error) {
	var zero T
	policy := c.cfg.PrimaryFollow.effective()
	v, err := fn()
	if err == nil || !policy.Enabled || !policy.RetryReads {
		return v, err
	}
	hint, followed, followErr := c.FollowPrimaryFromError(ctx, err, opts...)
	if followErr != nil {
		return zero, followErr
	}
	if !followed {
		return zero, err
	}
	v, retryErr := fn()
	if retryErr != nil {
		return zero, &PrimaryChangedRetryRequiredError{Hint: hint, Operation: operation, Err: retryErr}
	}
	return v, nil
}
func (c *Client) FollowPrimaryForUnsafe(ctx context.Context, operation string, err error, opts ...grpc.DialOption) error {
	hint, followed, followErr := c.FollowPrimaryFromError(ctx, err, opts...)
	if followErr != nil {
		return followErr
	}
	if !followed {
		return err
	}
	return &PrimaryChangedRetryRequiredError{Hint: hint, Operation: operation, Err: err}
}
