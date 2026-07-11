package mycel

import (
	"context"

	clientv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/client/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// ListDomains returns client-visible domains in a space.
func (c *Client) ListDomains(ctx context.Context, spaceID string, pageSize int32, pageToken string, includeSystem bool) (*clientv1.ListDomainsResponse, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	return c.Domain.ListDomains(callCtx, &clientv1.ListDomainsRequest{
		SpaceId:       spaceID,
		PageSize:      pageSize,
		PageToken:     pageToken,
		IncludeSystem: includeSystem,
	})
}

// GetDomain returns a domain by ID or key. If domainID is non-empty it is used;
// otherwise key is used.
func (c *Client) GetDomain(ctx context.Context, spaceID, domainID, key string) (*clientv1.Domain, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Domain.GetDomain(callCtx, &clientv1.GetDomainRequest{SpaceId: spaceID, DomainId: domainID, Key: key})
	if err != nil {
		return nil, err
	}
	return res.GetDomain(), nil
}

// CreateDomain creates a new non-system domain in a space.
func (c *Client) CreateDomain(ctx context.Context, spaceID, key, name, description string, discoveryMode clientv1.DomainDiscoveryMode) (*clientv1.Domain, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Domain.CreateDomain(callCtx, &clientv1.CreateDomainRequest{
		SpaceId:       spaceID,
		Key:           key,
		Name:          name,
		Description:   description,
		DiscoveryMode: discoveryMode,
	})
	if err != nil {
		return nil, err
	}
	return res.GetDomain(), nil
}

// UpdateDomainDiscoveryMode updates only the discovery mode for a domain.
func (c *Client) UpdateDomainDiscoveryMode(ctx context.Context, spaceID, domainID string, mode clientv1.DomainDiscoveryMode) (*clientv1.Domain, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Domain.UpdateDomain(callCtx, &clientv1.UpdateDomainRequest{
		SpaceId:  spaceID,
		DomainId: domainID,
		Domain: &clientv1.Domain{
			SpaceId:       spaceID,
			DomainId:      domainID,
			DiscoveryMode: mode,
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"discovery_mode"}},
	})
	if err != nil {
		return nil, err
	}
	return res.GetDomain(), nil
}
