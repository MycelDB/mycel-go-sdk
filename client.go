package mycel

import (
	"context"
	"fmt"

	adminv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/admin/v1"
	clientv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/client/v1"
	"google.golang.org/grpc"
)

type Client struct {
	Conn *grpc.ClientConn

	Auth         clientv1.AuthServiceClient
	Space        clientv1.SpaceServiceClient
	Domain       clientv1.DomainServiceClient
	Template     clientv1.TemplateServiceClient
	Session      clientv1.SessionServiceClient
	Transaction  clientv1.TransactionServiceClient
	Graph        clientv1.GraphServiceClient
	Blob         clientv1.BlobServiceClient
	Query        clientv1.QueryServiceClient
	ImportExport clientv1.ImportExportServiceClient
	Metadata     clientv1.MetadataCatalogServiceClient
	Semantic     clientv1.SemanticServiceClient
	ChangeStream clientv1.ChangeStreamServiceClient

	tokens *tokenSource
	cfg    Config
}

type AdminClient struct {
	Conn *grpc.ClientConn

	Auth                adminv1.AdminAuthServiceClient
	Operators           adminv1.AdminOperatorServiceClient
	Users               adminv1.AdminUserServiceClient
	Spaces              adminv1.AdminSpaceServiceClient
	Domains             adminv1.AdminDomainServiceClient
	Semantic            adminv1.AdminSemanticServiceClient
	SemanticMaintenance adminv1.AdminSemanticMaintenanceServiceClient
	SemanticMigration   adminv1.AdminSemanticMigrationServiceClient
	Inference           adminv1.AdminInferenceServiceClient
	Backup              adminv1.AdminBackupServiceClient

	tokens *tokenSource
	cfg    Config
}

func Dial(ctx context.Context, cfg Config, opts ...grpc.DialOption) (*Client, error) {
	tokens := newTokenSource(cfg.AccessToken)
	conn, err := dial(ctx, cfg, tokens, opts...)
	if err != nil {
		return nil, err
	}
	c := &Client{Conn: conn, tokens: tokens, cfg: cfg}
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
	if cfg.Username != "" || cfg.Password != "" {
		if _, err := c.Login(ctx, cfg.Username, cfg.Password); err != nil {
			_ = conn.Close()
			return nil, err
		}
	}
	return c, nil
}

func DialAdmin(ctx context.Context, cfg Config, opts ...grpc.DialOption) (*AdminClient, error) {
	tokens := newTokenSource(cfg.AccessToken)
	conn, err := dial(ctx, cfg, tokens, opts...)
	if err != nil {
		return nil, err
	}
	c := &AdminClient{Conn: conn, tokens: tokens, cfg: cfg}
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
	if cfg.Username != "" || cfg.Password != "" {
		if _, err := c.LoginOperator(ctx, cfg.Username, cfg.Password); err != nil {
			_ = conn.Close()
			return nil, err
		}
	}
	return c, nil
}

func dial(ctx context.Context, cfg Config, tokens *tokenSource, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	transport, err := transportOption(cfg)
	if err != nil {
		return nil, err
	}
	dialOpts := []grpc.DialOption{
		transport,
		grpc.WithUnaryInterceptor(tokens.unaryInterceptor),
		grpc.WithStreamInterceptor(tokens.streamInterceptor),
	}
	dialOpts = append(dialOpts, opts...)
	conn, err := grpc.NewClient(cfg.addr(), dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("dial myceld %s: %w", cfg.addr(), err)
	}
	if ctx != nil {
		select {
		case <-ctx.Done():
			_ = conn.Close()
			return nil, ctx.Err()
		default:
		}
	}
	return conn, nil
}

func (c *Client) Close() error {
	if c == nil || c.Conn == nil {
		return nil
	}
	return c.Conn.Close()
}

func (c *AdminClient) Close() error {
	if c == nil || c.Conn == nil {
		return nil
	}
	return c.Conn.Close()
}
