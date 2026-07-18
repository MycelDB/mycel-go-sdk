package mycel

import (
	"os"
	"strings"
	"time"
)

const DefaultAddr = "127.0.0.1:9091"

type Config struct {
	Addr string

	Username              string
	Password              string
	AccessToken           string
	AccessTokenExpireTime time.Time
	RefreshToken          string
	RefreshBefore         time.Duration
	CallTimeout           time.Duration
	PrimaryFollow         PrimaryFollowPolicy

	TLS                   bool
	TLSCAFile             string
	TLSServerName         string
	TLSInsecureSkipVerify bool
	TLSClientCertFile     string
	TLSClientKeyFile      string

	ClientName    string
	ClientVersion string
	Platform      string
	DeviceLabel   string
}

type PrimaryFollowPolicy struct {
	Enabled      bool
	RetryReads   bool
	RetryUnsafe  bool
	MaxRedirects int
}

func DefaultPrimaryFollowPolicy() PrimaryFollowPolicy {
	return PrimaryFollowPolicy{Enabled: true, RetryReads: true, MaxRedirects: 1}
}

func (p PrimaryFollowPolicy) effective() PrimaryFollowPolicy {
	if p.MaxRedirects <= 0 {
		p.MaxRedirects = 1
	}
	return p
}

func (c Config) addr() string {
	if strings.TrimSpace(c.Addr) != "" {
		return strings.TrimSpace(c.Addr)
	}
	return DefaultAddr
}

func ConfigFromEnv() Config {
	return Config{
		Addr:                  firstNonEmpty(os.Getenv("MYCELD_GRPC_ADDR"), DefaultAddr),
		Username:              os.Getenv("MYCEL_USERNAME"),
		Password:              os.Getenv("MYCEL_PASSWORD"),
		AccessToken:           os.Getenv("MYCEL_ACCESS_TOKEN"),
		AccessTokenExpireTime: parseTime(os.Getenv("MYCEL_ACCESS_TOKEN_EXPIRE_TIME")),
		RefreshToken:          os.Getenv("MYCEL_REFRESH_TOKEN"),
		RefreshBefore:         parseDuration(os.Getenv("MYCEL_REFRESH_BEFORE")),
		CallTimeout:           parseDuration(os.Getenv("MYCEL_CALL_TIMEOUT")),
		PrimaryFollow:         DefaultPrimaryFollowPolicy(),
		TLS:                   parseBool(os.Getenv("MYCELD_TLS")),
		TLSCAFile:             os.Getenv("MYCELD_TLS_CA_FILE"),
		TLSServerName:         os.Getenv("MYCELD_TLS_SERVER_NAME"),
		TLSInsecureSkipVerify: parseBool(os.Getenv("MYCELD_TLS_INSECURE_SKIP_VERIFY")),
		TLSClientCertFile:     os.Getenv("MYCELD_TLS_CLIENT_CERT_FILE"),
		TLSClientKeyFile:      os.Getenv("MYCELD_TLS_CLIENT_KEY_FILE"),
		ClientName:            firstNonEmpty(os.Getenv("MYCEL_CLIENT_NAME"), "mycel-go-sdk"),
		ClientVersion:         os.Getenv("MYCEL_CLIENT_VERSION"),
		Platform:              firstNonEmpty(os.Getenv("MYCEL_CLIENT_PLATFORM"), "go"),
		DeviceLabel:           os.Getenv("MYCEL_CLIENT_DEVICE_LABEL"),
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func parseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "t", "true", "y", "yes", "on":
		return true
	default:
		return false
	}
}

func parseDuration(value string) time.Duration {
	if strings.TrimSpace(value) == "" {
		return 0
	}
	d, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0
	}
	return d
}

func parseTime(value string) time.Time {
	if strings.TrimSpace(value) == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}
	}
	return parsed.UTC()
}
