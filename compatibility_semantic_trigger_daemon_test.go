package mycel

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	adminv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func TestV021SemanticTriggerCompatibilityAgainstDaemon(t *testing.T) {
	spaceID := strings.TrimSpace(os.Getenv("MYCEL_V021_COMPAT_SPACE_ID"))
	domainID := strings.TrimSpace(os.Getenv("MYCEL_V021_COMPAT_DOMAIN_ID"))
	if spaceID == "" || domainID == "" {
		t.Skip("set MYCEL_V021_COMPAT_SPACE_ID and MYCEL_V021_COMPAT_DOMAIN_ID to run the read-only daemon compatibility check")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	admin, err := DialAdmin(ctx, ConfigFromEnv())
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()

	authCtx := admin.AuthContext(ctx)
	rule := &adminv1.SemanticGenerationRule{
		SpaceId:  spaceID,
		DomainId: domainID,
		Key:      "sdk-v021-semantic-trigger-compatibility",
		Trigger: &adminv1.SemanticTriggerPolicy{Events: []adminv1.SemanticTriggerEventType{
			adminv1.SemanticTriggerEventType_SEMANTIC_TRIGGER_EVENT_TYPE_NODE_CREATED,
			adminv1.SemanticTriggerEventType_SEMANTIC_TRIGGER_EVENT_TYPE_NODE_UPDATED,
			adminv1.SemanticTriggerEventType_SEMANTIC_TRIGGER_EVENT_TYPE_EDGE_CREATED,
			adminv1.SemanticTriggerEventType_SEMANTIC_TRIGGER_EVENT_TYPE_EDGE_DELETED,
		}},
		Selector: &adminv1.SemanticTargetSelector{Mode: "node_type", Labels: []string{"CompatibilityFixture"}},
		Source:   &adminv1.SemanticSourceAssemblyPolicy{Mode: "self"},
		Embeddings: []*adminv1.SemanticEmbeddingBinding{{
			Key:                 "default",
			Purpose:             "compatibility",
			IntelligenceProfile: "compatibility-profile",
			VectorStore:         "compatibility-vector-store",
			Enabled:             true,
		}},
	}
	for i, event := range rule.GetTrigger().GetEvents() {
		if _, known := adminv1.SemanticTriggerEventType_name[int32(event)]; !known || event == adminv1.SemanticTriggerEventType_SEMANTIC_TRIGGER_EVENT_TYPE_UNSPECIFIED {
			t.Fatalf("trigger.events[%d] = %d is not a supported explicit v0.21 enum", i, event)
		}
	}

	validation, err := admin.Semantic.ValidateSemanticRule(authCtx, &adminv1.ValidateSemanticRuleRequest{Rule: rule})
	if err != nil {
		t.Fatalf("validate typed v0.21 semantic rule: %v", err)
	}
	if !validation.GetValid() {
		t.Fatalf("typed v0.21 semantic rule is invalid: %v", validation.GetDiagnostics())
	}

	legacyEquivalent := proto.Clone(rule).(*adminv1.SemanticGenerationRule)
	legacyEquivalent.Trigger.Events = []adminv1.SemanticTriggerEventType{adminv1.SemanticTriggerEventType(110)}
	_, err = admin.Semantic.ValidateSemanticRule(authCtx, &adminv1.ValidateSemanticRuleRequest{Rule: legacyEquivalent})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("legacy-equivalent request error = %v, want InvalidArgument", err)
	}
	const wantDiagnostic = "trigger.events[0] contains unsupported semantic trigger event value 110"
	if !strings.Contains(status.Convert(err).Message(), wantDiagnostic) {
		t.Fatalf("legacy-equivalent diagnostic = %q, want %q", status.Convert(err).Message(), wantDiagnostic)
	}
}
