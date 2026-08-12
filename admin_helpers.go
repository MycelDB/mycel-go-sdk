package mycel

import (
	"context"
	"fmt"
	"strings"
	"time"

	adminv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/admin/v1"
	commonv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/common/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OperatorInfo struct {
	OperatorID string
	Username   string
}

type PrincipalInfo struct {
	UserID      string
	PrincipalID string
	Username    string
}

type UserInfo struct {
	UserID   string
	Username string
	State    string
}

type UserSessionInfo struct {
	UserID                string
	Username              string
	AccessToken           string
	AccessTokenExpireTime time.Time
	RefreshToken          string
	AuthSessionID         string
}

type SpaceInfo struct {
	SpaceID string
	Name    string
}

type SpaceGrantInfo struct {
	GrantID     string
	SpaceID     string
	UserID      string
	PrincipalID string
	Role        string
}

type DomainInfo struct {
	SpaceID  string
	DomainID string
	Key      string
	Name     string
	Default  bool
	System   bool
}

func (c *AdminClient) WhoAmI(ctx context.Context) (*OperatorInfo, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Auth.WhoAmI(callCtx, &commonv1.WhoAmIRequest{})
	if err != nil {
		return nil, err
	}
	principal := res.GetPrincipal()
	return &OperatorInfo{OperatorID: principal.GetPrincipalId(), Username: principal.GetUsername()}, nil
}

func (c *AdminClient) FindUser(ctx context.Context, username string) (*UserInfo, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Principals.FindPrincipal(callCtx, &adminv1.FindPrincipalRequest{Lookup: &adminv1.FindPrincipalRequest_Username{Username: strings.TrimSpace(username)}})
	if err != nil {
		return nil, err
	}
	return userInfo(res.GetPrincipal()), nil
}

func (c *AdminClient) EnsureUser(ctx context.Context, username, password string) (*UserInfo, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	user, err := c.FindUser(ctx, username)
	if err == nil {
		return user, nil
	}
	if status.Code(err) != codes.NotFound {
		return nil, err
	}
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Principals.CreatePrincipal(callCtx, &adminv1.CreatePrincipalRequest{Username: username, Password: &password, Type: commonv1.PrincipalType_PRINCIPAL_TYPE_HUMAN, LoginEnabled: true})
	if err != nil {
		return nil, err
	}
	return userInfo(res.GetPrincipal()), nil
}

func (c *AdminClient) CreateUserSession(ctx context.Context, userID string) (*UserSessionInfo, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Principals.CreatePrincipalSession(callCtx, &adminv1.CreatePrincipalSessionRequest{PrincipalId: userID, Client: c.adminClientInfo()})
	if err != nil {
		return nil, err
	}
	user := userInfo(res.GetPrincipal())
	return &UserSessionInfo{UserID: user.UserID, Username: user.Username, AccessToken: res.GetAccessToken(), AccessTokenExpireTime: timestampAsTime(res.GetAccessTokenExpireTime()), RefreshToken: res.GetRefreshToken(), AuthSessionID: res.GetAuthSessionId()}, nil
}

func (c *AdminClient) DialUserClient(ctx context.Context, userID string) (*Client, *UserSessionInfo, error) {
	session, err := c.CreateUserSession(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	cfg := c.cfg
	cfg.Username = ""
	cfg.Password = ""
	cfg.AccessToken = session.AccessToken
	cfg.AccessTokenExpireTime = session.AccessTokenExpireTime
	cfg.RefreshToken = session.RefreshToken
	client, err := Dial(ctx, cfg)
	if err != nil {
		_ = c.RevokeUserSession(ctx, session.UserID, session.AuthSessionID)
		return nil, nil, err
	}
	return client, session, nil
}

func (c *AdminClient) RevokeUserSession(ctx context.Context, userID, authSessionID string) error {
	userID = strings.TrimSpace(userID)
	authSessionID = strings.TrimSpace(authSessionID)
	if userID == "" || authSessionID == "" {
		return nil
	}
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	_, err := c.Principals.RevokePrincipalSession(callCtx, &adminv1.RevokePrincipalSessionRequest{PrincipalId: userID, AuthSessionId: authSessionID})
	return err
}

func (c *AdminClient) FindSpaceByName(ctx context.Context, name string) (*SpaceInfo, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("space name is required")
	}
	var token string
	var found *SpaceInfo
	for {
		callCtx, cancel := c.AuthCallContext(ctx)
		res, err := c.Spaces.ListSpaces(callCtx, &adminv1.AdminSpaceServiceListSpacesRequest{PageSize: 100, PageToken: token})
		cancel()
		if err != nil {
			return nil, err
		}
		for _, sp := range res.GetSpaces() {
			if sp.GetName() != name {
				continue
			}
			info := &SpaceInfo{SpaceID: sp.GetSpaceId(), Name: sp.GetName()}
			if found != nil {
				return nil, fmt.Errorf("multiple spaces named %q", name)
			}
			found = info
		}
		if res.GetNextPageToken() == "" {
			break
		}
		token = res.GetNextPageToken()
	}
	if found == nil {
		return nil, status.Errorf(codes.NotFound, "space %q not found", name)
	}
	return found, nil
}

func (c *AdminClient) EnsureSpace(ctx context.Context, name, ownerUsername, defaultDomainKey, defaultDomainName string) (*SpaceInfo, *DomainInfo, error) {
	if sp, err := c.FindSpaceByName(ctx, name); err == nil {
		domain, domainErr := c.GetDomain(ctx, sp.SpaceID, firstNonEmpty(defaultDomainKey, "default"))
		if domainErr != nil {
			return sp, nil, domainErr
		}
		return sp, domain, nil
	} else if status.Code(err) != codes.NotFound {
		return nil, nil, err
	}

	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Spaces.CreateSpace(callCtx, &adminv1.CreateSpaceRequest{
		Name:              strings.TrimSpace(name),
		OwnerUsername:     strings.TrimSpace(ownerUsername),
		DefaultDomainKey:  firstNonEmpty(defaultDomainKey, "default"),
		DefaultDomainName: firstNonEmpty(defaultDomainName, defaultDomainKey, "default"),
	})
	if err != nil {
		return nil, nil, err
	}
	sp := &SpaceInfo{SpaceID: res.GetSpace().GetSpaceId(), Name: res.GetSpace().GetName()}
	domain := &DomainInfo{SpaceID: sp.SpaceID, DomainID: res.GetDefaultDomainId(), Key: firstNonEmpty(defaultDomainKey, "default"), Name: firstNonEmpty(defaultDomainName, defaultDomainKey, "default"), Default: true}
	return sp, domain, nil
}

func (c *AdminClient) GrantSpaceUser(ctx context.Context, spaceID, userID, role string) (*SpaceGrantInfo, error) {
	return c.GrantSpacePrincipal(ctx, spaceID, userID, role)
}

func (c *AdminClient) GrantSpacePrincipal(ctx context.Context, spaceID, principalID, role string) (*SpaceGrantInfo, error) {
	spaceID = strings.TrimSpace(spaceID)
	principalID = strings.TrimSpace(principalID)
	if spaceID == "" || principalID == "" {
		return nil, fmt.Errorf("space id and principal id are required")
	}
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Spaces.GrantSpacePrincipal(callCtx, &adminv1.GrantSpacePrincipalRequest{SpaceId: spaceID, PrincipalId: principalID, Role: sdkSpaceRole(role)})
	if err != nil {
		return nil, err
	}
	return spaceGrantInfo(res.GetGrant()), nil
}

func (c *AdminClient) GetDomain(ctx context.Context, spaceID, domainRef string) (*DomainInfo, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Domains.GetDomain(callCtx, &adminv1.AdminDomainServiceGetDomainRequest{SpaceId: spaceID, DomainRef: domainRef})
	if err != nil {
		return nil, err
	}
	d := res.GetDomain()
	return &DomainInfo{SpaceID: d.GetSpaceId(), DomainID: d.GetDomainId(), Key: d.GetKey(), Name: d.GetName(), Default: d.GetDefault(), System: d.GetSystem()}, nil
}

func sdkSpaceRole(role string) commonv1.SpaceRole {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin":
		return commonv1.SpaceRole_SPACE_ROLE_ADMIN
	case "writer", "write":
		return commonv1.SpaceRole_SPACE_ROLE_WRITER
	case "reader", "read":
		return commonv1.SpaceRole_SPACE_ROLE_READER
	default:
		return commonv1.SpaceRole_SPACE_ROLE_UNSPECIFIED
	}
}

func spaceGrantInfo(grant *commonv1.AccessGrant) *SpaceGrantInfo {
	if grant == nil {
		return &SpaceGrantInfo{}
	}
	role := ""
	if len(grant.GetRoles()) > 0 {
		switch grant.GetRoles()[0] {
		case commonv1.SpaceRole_SPACE_ROLE_ADMIN:
			role = "admin"
		case commonv1.SpaceRole_SPACE_ROLE_WRITER:
			role = "writer"
		case commonv1.SpaceRole_SPACE_ROLE_READER:
			role = "reader"
		}
	}
	spaceID := ""
	if grant.GetScope() != nil {
		spaceID = grant.GetScope().GetSpaceId()
	}
	principalID := grant.GetPrincipal().GetId()
	return &SpaceGrantInfo{GrantID: grant.GetAccessGrantId(), SpaceID: spaceID, UserID: principalID, PrincipalID: principalID, Role: role}
}

func userInfo(p *adminv1.Principal) *UserInfo {
	if p == nil {
		return &UserInfo{}
	}
	return &UserInfo{UserID: p.GetPrincipalId(), Username: p.GetUsername(), State: p.GetState().String()}
}
