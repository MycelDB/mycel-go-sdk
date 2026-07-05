package mycel

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func transportOption(cfg Config) (grpc.DialOption, error) {
	if !cfg.TLS {
		return grpc.WithTransportCredentials(insecure.NewCredentials()), nil
	}
	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: cfg.TLSServerName, InsecureSkipVerify: cfg.TLSInsecureSkipVerify} //nolint:gosec // explicit SDK config for local/testing parity with CLI
	if cfg.TLSCAFile != "" {
		pem, err := os.ReadFile(cfg.TLSCAFile)
		if err != nil {
			return nil, fmt.Errorf("read TLS CA file: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("TLS CA file contains no certificates")
		}
		tlsCfg.RootCAs = pool
	}
	if (cfg.TLSClientCertFile == "") != (cfg.TLSClientKeyFile == "") {
		return nil, fmt.Errorf("TLS client cert and key must be set together")
	}
	if cfg.TLSClientCertFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.TLSClientCertFile, cfg.TLSClientKeyFile)
		if err != nil {
			return nil, fmt.Errorf("load TLS client certificate: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}
	return grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)), nil
}
