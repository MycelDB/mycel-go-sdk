package mycel

import (
	"context"
	"testing"

	clientv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/client/v1"
	"google.golang.org/grpc"
)

type fakeDomainServiceClient struct {
	clientv1.DomainServiceClient
	createReq *clientv1.CreateDomainRequest
	updateReq *clientv1.UpdateDomainRequest
	getReq    *clientv1.GetDomainRequest
	listReq   *clientv1.ListDomainsRequest
}

func (f *fakeDomainServiceClient) ListDomains(ctx context.Context, in *clientv1.ListDomainsRequest, opts ...grpc.CallOption) (*clientv1.ListDomainsResponse, error) {
	f.listReq = in
	return &clientv1.ListDomainsResponse{Domains: []*clientv1.Domain{{SpaceId: in.GetSpaceId()}}}, nil
}

func (f *fakeDomainServiceClient) GetDomain(ctx context.Context, in *clientv1.GetDomainRequest, opts ...grpc.CallOption) (*clientv1.GetDomainResponse, error) {
	f.getReq = in
	return &clientv1.GetDomainResponse{Domain: &clientv1.Domain{SpaceId: in.GetSpaceId(), DomainId: in.GetDomainId(), Key: in.GetKey()}}, nil
}

func (f *fakeDomainServiceClient) CreateDomain(ctx context.Context, in *clientv1.CreateDomainRequest, opts ...grpc.CallOption) (*clientv1.CreateDomainResponse, error) {
	f.createReq = in
	return &clientv1.CreateDomainResponse{Domain: &clientv1.Domain{SpaceId: in.GetSpaceId(), Key: in.GetKey(), Name: in.GetName(), Description: in.GetDescription(), DiscoveryMode: in.GetDiscoveryMode()}}, nil
}

func (f *fakeDomainServiceClient) UpdateDomain(ctx context.Context, in *clientv1.UpdateDomainRequest, opts ...grpc.CallOption) (*clientv1.UpdateDomainResponse, error) {
	f.updateReq = in
	return &clientv1.UpdateDomainResponse{Domain: in.GetDomain()}, nil
}

func TestCreateDomainHelperBuildsRequest(t *testing.T) {
	fake := &fakeDomainServiceClient{}
	c := &Client{Domain: fake, tokens: newTokenSource("tok")}

	domain, err := c.CreateDomain(context.Background(), "space-1", "domain-key", "Domain Name", "description", clientv1.DomainDiscoveryMode_DOMAIN_DISCOVERY_MODE_DIRECT_ONLY)
	if err != nil {
		t.Fatalf("CreateDomain() error = %v", err)
	}
	if fake.createReq == nil {
		t.Fatal("expected CreateDomain request")
	}
	if fake.createReq.GetSpaceId() != "space-1" || fake.createReq.GetKey() != "domain-key" || fake.createReq.GetName() != "Domain Name" || fake.createReq.GetDescription() != "description" || fake.createReq.GetDiscoveryMode() != clientv1.DomainDiscoveryMode_DOMAIN_DISCOVERY_MODE_DIRECT_ONLY {
		t.Fatalf("unexpected CreateDomain request: %+v", fake.createReq)
	}
	if domain.GetDiscoveryMode() != clientv1.DomainDiscoveryMode_DOMAIN_DISCOVERY_MODE_DIRECT_ONLY {
		t.Fatalf("unexpected returned domain: %+v", domain)
	}
}

func TestUpdateDomainDiscoveryModeHelperBuildsRequest(t *testing.T) {
	fake := &fakeDomainServiceClient{}
	c := &Client{Domain: fake, tokens: newTokenSource("tok")}

	domain, err := c.UpdateDomainDiscoveryMode(context.Background(), "space-1", "domain-1", clientv1.DomainDiscoveryMode_DOMAIN_DISCOVERY_MODE_NORMAL)
	if err != nil {
		t.Fatalf("UpdateDomainDiscoveryMode() error = %v", err)
	}
	if fake.updateReq == nil {
		t.Fatal("expected UpdateDomain request")
	}
	if fake.updateReq.GetSpaceId() != "space-1" || fake.updateReq.GetDomainId() != "domain-1" {
		t.Fatalf("unexpected UpdateDomain request IDs: %+v", fake.updateReq)
	}
	if fake.updateReq.GetDomain().GetSpaceId() != "space-1" || fake.updateReq.GetDomain().GetDomainId() != "domain-1" || fake.updateReq.GetDomain().GetDiscoveryMode() != clientv1.DomainDiscoveryMode_DOMAIN_DISCOVERY_MODE_NORMAL {
		t.Fatalf("unexpected UpdateDomain resource: %+v", fake.updateReq.GetDomain())
	}
	if fake.updateReq.GetUpdateMask() == nil || len(fake.updateReq.GetUpdateMask().GetPaths()) != 1 || fake.updateReq.GetUpdateMask().GetPaths()[0] != "discovery_mode" {
		t.Fatalf("unexpected update mask: %+v", fake.updateReq.GetUpdateMask())
	}
	if domain.GetDiscoveryMode() != clientv1.DomainDiscoveryMode_DOMAIN_DISCOVERY_MODE_NORMAL {
		t.Fatalf("unexpected returned domain: %+v", domain)
	}
}
