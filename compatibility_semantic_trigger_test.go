package mycel

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	adminv1 "github.com/myceldb/mycel-go-sdk/gen/go/mycel/admin/v1"
	"google.golang.org/protobuf/proto"
)

func TestV021SemanticTriggerTypedWireContract(t *testing.T) {
	policy := &adminv1.SemanticTriggerPolicy{Events: []adminv1.SemanticTriggerEventType{
		adminv1.SemanticTriggerEventType_SEMANTIC_TRIGGER_EVENT_TYPE_CHANGED,
		adminv1.SemanticTriggerEventType_SEMANTIC_TRIGGER_EVENT_TYPE_NODE_CREATED,
		adminv1.SemanticTriggerEventType_SEMANTIC_TRIGGER_EVENT_TYPE_NODE_UPDATED,
		adminv1.SemanticTriggerEventType_SEMANTIC_TRIGGER_EVENT_TYPE_NODE_DELETED,
		adminv1.SemanticTriggerEventType_SEMANTIC_TRIGGER_EVENT_TYPE_EDGE_CHANGED,
		adminv1.SemanticTriggerEventType_SEMANTIC_TRIGGER_EVENT_TYPE_EDGE_CREATED,
		adminv1.SemanticTriggerEventType_SEMANTIC_TRIGGER_EVENT_TYPE_EDGE_UPDATED,
		adminv1.SemanticTriggerEventType_SEMANTIC_TRIGGER_EVENT_TYPE_EDGE_DELETED,
	}}

	got, err := proto.Marshal(policy)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0x0a, 0x08, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	if !bytes.Equal(got, want) {
		t.Fatalf("typed trigger wire = %x, want %x", got, want)
	}

	var decoded adminv1.SemanticTriggerPolicy
	if err := proto.Unmarshal(got, &decoded); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(policy, &decoded) {
		t.Fatalf("decoded policy = %#v, want %#v", &decoded, policy)
	}
}

func TestV020LegacySemanticTriggerWireFixtureDecodesAsUnsupportedEnum(t *testing.T) {
	hexFixture, err := os.ReadFile("testdata/semantic-trigger-v020-node-created.hex")
	if err != nil {
		t.Fatal(err)
	}
	wire, err := hex.DecodeString(string(bytes.TrimSpace(hexFixture)))
	if err != nil {
		t.Fatal(err)
	}

	var decoded adminv1.SemanticTriggerPolicy
	if err := proto.Unmarshal(wire, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.GetEvents()) == 0 {
		t.Fatal("legacy field-1 string payload decoded without event values")
	}
	if got := decoded.GetEvents()[0]; got != adminv1.SemanticTriggerEventType(110) {
		t.Fatalf("first decoded event = %d, want 110 (ASCII 'n')", got)
	}
	if _, known := adminv1.SemanticTriggerEventType_name[int32(decoded.GetEvents()[0])]; known {
		t.Fatalf("legacy event value %d unexpectedly became a known v0.21 enum", decoded.GetEvents()[0])
	}

	const wantDiagnostic = "trigger.events[0] contains unsupported semantic trigger event value 110"
	gotDiagnostic := fmt.Sprintf("trigger.events[0] contains unsupported semantic trigger event value %d", decoded.GetEvents()[0])
	if gotDiagnostic != wantDiagnostic {
		t.Fatalf("legacy diagnostic = %q, want %q", gotDiagnostic, wantDiagnostic)
	}
}
