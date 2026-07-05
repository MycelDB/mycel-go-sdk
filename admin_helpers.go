package mycel

import (
	"context"
	"fmt"
	"strings"

	adminv1 "github.com/myceldb/mycel-api/gen/go/mycel/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type OperatorInfo struct {
	OperatorID string
	Username   string
}

type PrincipalInfo struct {
	UserID   string
	Username string
}

type UserInfo struct {
	UserID   string
	Username string
	State    string
}

type SpaceInfo struct {
	SpaceID string
	Name    string
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
	res, err := c.Auth.WhoAmI(callCtx, &adminv1.WhoAmIRequest{})
	if err != nil {
		return nil, err
	}
	op := res.GetOperator()
	return &OperatorInfo{OperatorID: op.GetOperatorId(), Username: op.GetUsername()}, nil
}

func (c *AdminClient) FindUser(ctx context.Context, username string) (*UserInfo, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Users.FindUser(callCtx, &adminv1.FindUserRequest{Username: strings.TrimSpace(username)})
	if err != nil {
		return nil, err
	}
	return userInfo(res.GetUser()), nil
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
	res, err := c.Users.CreateUser(callCtx, &adminv1.CreateUserRequest{Username: username, Password: proto.String(password)})
	if err != nil {
		return nil, err
	}
	return userInfo(res.GetUser()), nil
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

func userInfo(u *adminv1.User) *UserInfo {
	if u == nil {
		return &UserInfo{}
	}
	return &UserInfo{UserID: u.GetUserId(), Username: u.GetUsername(), State: u.GetState().String()}
}
