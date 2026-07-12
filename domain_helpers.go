package mycel

import (
	"context"

	clientv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/client/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// DomainPolicy controls discovery, search, semantic, and write behavior for a domain.
type DomainPolicy struct {
	DiscoveryMode clientv1.DomainDiscoveryMode
	SearchMode    clientv1.DomainSearchMode
	SemanticMode  clientv1.DomainSemanticMode
	ReadOnly      bool
}

// NormalDomainPolicy returns the default writable, discoverable, searchable domain policy.
func NormalDomainPolicy() DomainPolicy {
	return DomainPolicy{DiscoveryMode: clientv1.DomainDiscoveryMode_DOMAIN_DISCOVERY_MODE_NORMAL, SearchMode: clientv1.DomainSearchMode_DOMAIN_SEARCH_MODE_NORMAL, SemanticMode: clientv1.DomainSemanticMode_DOMAIN_SEMANTIC_MODE_NORMAL}
}

// PrivateDomainPolicy returns the policy for directly addressed private data such as steward profiles.
func PrivateDomainPolicy() DomainPolicy {
	return DomainPolicy{DiscoveryMode: clientv1.DomainDiscoveryMode_DOMAIN_DISCOVERY_MODE_EXPLICIT_ONLY, SearchMode: clientv1.DomainSearchMode_DOMAIN_SEARCH_MODE_DISABLED, SemanticMode: clientv1.DomainSemanticMode_DOMAIN_SEMANTIC_MODE_DISABLED}
}

// ManualDomainPolicy returns the policy for read-only manual/help domains that are searchable only when explicitly targeted.
func ManualDomainPolicy() DomainPolicy {
	return DomainPolicy{DiscoveryMode: clientv1.DomainDiscoveryMode_DOMAIN_DISCOVERY_MODE_EXPLICIT_ONLY, SearchMode: clientv1.DomainSearchMode_DOMAIN_SEARCH_MODE_EXPLICIT_ONLY, SemanticMode: clientv1.DomainSemanticMode_DOMAIN_SEMANTIC_MODE_EXPLICIT_ONLY, ReadOnly: true}
}

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

// CreateDomain creates a new non-system domain in a space using the supplied policy.
func (c *Client) CreateDomain(ctx context.Context, spaceID, key, name, description string, policy DomainPolicy) (*clientv1.Domain, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Domain.CreateDomain(callCtx, &clientv1.CreateDomainRequest{
		SpaceId:       spaceID,
		Key:           key,
		Name:          name,
		Description:   description,
		DiscoveryMode: policy.DiscoveryMode,
		SearchMode:    policy.SearchMode,
		SemanticMode:  policy.SemanticMode,
		ReadOnly:      policy.ReadOnly,
	})
	if err != nil {
		return nil, err
	}
	return res.GetDomain(), nil
}

// UpdateDomainPolicy updates the discovery/search/semantic/read-only policy for a domain.
func (c *Client) UpdateDomainPolicy(ctx context.Context, spaceID, domainID string, policy DomainPolicy) (*clientv1.Domain, error) {
	callCtx, cancel := c.AuthCallContext(ctx)
	defer cancel()
	res, err := c.Domain.UpdateDomain(callCtx, &clientv1.UpdateDomainRequest{
		SpaceId:  spaceID,
		DomainId: domainID,
		Domain: &clientv1.Domain{
			SpaceId:       spaceID,
			DomainId:      domainID,
			DiscoveryMode: policy.DiscoveryMode,
			SearchMode:    policy.SearchMode,
			SemanticMode:  policy.SemanticMode,
			ReadOnly:      policy.ReadOnly,
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"discovery_mode", "search_mode", "semantic_mode", "read_only"}},
	})
	if err != nil {
		return nil, err
	}
	return res.GetDomain(), nil
}
