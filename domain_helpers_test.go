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
	return &clientv1.CreateDomainResponse{Domain: &clientv1.Domain{SpaceId: in.GetSpaceId(), Key: in.GetKey(), Name: in.GetName(), Description: in.GetDescription(), DiscoveryMode: in.GetDiscoveryMode(), SearchMode: in.GetSearchMode(), SemanticMode: in.GetSemanticMode(), ReadOnly: in.GetReadOnly()}}, nil
}

func (f *fakeDomainServiceClient) UpdateDomain(ctx context.Context, in *clientv1.UpdateDomainRequest, opts ...grpc.CallOption) (*clientv1.UpdateDomainResponse, error) {
	f.updateReq = in
	return &clientv1.UpdateDomainResponse{Domain: in.GetDomain()}, nil
}

func TestCreateDomainHelperBuildsPolicyRequest(t *testing.T) {
	fake := &fakeDomainServiceClient{}
	c := &Client{Domain: fake, tokens: newTokenSource("tok")}
	policy := ManualDomainPolicy()

	domain, err := c.CreateDomain(context.Background(), "space-1", "domain-key", "Domain Name", "description", policy)
	if err != nil {
		t.Fatalf("CreateDomain() error = %v", err)
	}
	if fake.createReq == nil {
		t.Fatal("expected CreateDomain request")
	}
	if fake.createReq.GetSpaceId() != "space-1" || fake.createReq.GetKey() != "domain-key" || fake.createReq.GetName() != "Domain Name" || fake.createReq.GetDescription() != "description" {
		t.Fatalf("unexpected CreateDomain request: %+v", fake.createReq)
	}
	if fake.createReq.GetDiscoveryMode() != policy.DiscoveryMode || fake.createReq.GetSearchMode() != policy.SearchMode || fake.createReq.GetSemanticMode() != policy.SemanticMode || fake.createReq.GetReadOnly() != policy.ReadOnly {
		t.Fatalf("unexpected CreateDomain policy: %+v", fake.createReq)
	}
	if domain.GetDiscoveryMode() != policy.DiscoveryMode || domain.GetSearchMode() != policy.SearchMode || domain.GetSemanticMode() != policy.SemanticMode || domain.GetReadOnly() != policy.ReadOnly {
		t.Fatalf("unexpected returned domain: %+v", domain)
	}
}

func TestUpdateDomainPolicyHelperBuildsRequest(t *testing.T) {
	fake := &fakeDomainServiceClient{}
	c := &Client{Domain: fake, tokens: newTokenSource("tok")}
	policy := PrivateDomainPolicy()

	domain, err := c.UpdateDomainPolicy(context.Background(), "space-1", "domain-1", policy)
	if err != nil {
		t.Fatalf("UpdateDomainPolicy() error = %v", err)
	}
	if fake.updateReq == nil {
		t.Fatal("expected UpdateDomain request")
	}
	if fake.updateReq.GetSpaceId() != "space-1" || fake.updateReq.GetDomainId() != "domain-1" {
		t.Fatalf("unexpected UpdateDomain request IDs: %+v", fake.updateReq)
	}
	if fake.updateReq.GetDomain().GetSpaceId() != "space-1" || fake.updateReq.GetDomain().GetDomainId() != "domain-1" || fake.updateReq.GetDomain().GetDiscoveryMode() != policy.DiscoveryMode || fake.updateReq.GetDomain().GetSearchMode() != policy.SearchMode || fake.updateReq.GetDomain().GetSemanticMode() != policy.SemanticMode || fake.updateReq.GetDomain().GetReadOnly() != policy.ReadOnly {
		t.Fatalf("unexpected UpdateDomain resource: %+v", fake.updateReq.GetDomain())
	}
	paths := fake.updateReq.GetUpdateMask().GetPaths()
	want := []string{"discovery_mode", "search_mode", "semantic_mode", "read_only"}
	if len(paths) != len(want) {
		t.Fatalf("unexpected update mask: %+v", fake.updateReq.GetUpdateMask())
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Fatalf("unexpected update mask: %+v", fake.updateReq.GetUpdateMask())
		}
	}
	if domain.GetDiscoveryMode() != policy.DiscoveryMode || domain.GetSearchMode() != policy.SearchMode || domain.GetSemanticMode() != policy.SemanticMode || domain.GetReadOnly() != policy.ReadOnly {
		t.Fatalf("unexpected returned domain: %+v", domain)
	}
}
