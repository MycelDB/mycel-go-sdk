package mycel

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

type compatibilityManifest struct {
	SchemaVersion              int                         `json:"schema_version"`
	Component                  string                      `json:"component"`
	ComponentRelease           string                      `json:"component_release"`
	SourceRevision             string                      `json:"source_revision"`
	MycelAPILine               string                      `json:"mycel_api_line"`
	SupportedDaemonAPILines    []string                    `json:"supported_daemon_api_lines"`
	IncompatibleDaemonAPILines []incompatibleDaemonAPILine `json:"incompatible_daemon_api_lines"`
	PublishedAt                string                      `json:"published_at"`
}

type incompatibleDaemonAPILine struct {
	Line      string   `json:"line"`
	Contracts []string `json:"contracts"`
}

func TestCompatibilityManifestV1(t *testing.T) {
	data, err := os.ReadFile(".release/mycel-compatibility.json")
	if err != nil {
		t.Fatal(err)
	}

	var manifest compatibilityManifest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		t.Fatal(err)
	}

	if manifest.SchemaVersion != 1 {
		t.Fatalf("schema_version = %d, want 1", manifest.SchemaVersion)
	}
	if manifest.Component != "mycel-go-sdk" {
		t.Fatalf("component = %q, want mycel-go-sdk", manifest.Component)
	}
	if manifest.ComponentRelease != "v0.24.0" {
		t.Fatalf("component_release = %q, want v0.24.0", manifest.ComponentRelease)
	}
	if manifest.SourceRevision != "0c3004700495a1c4c40b704a47216cbca4ac11f3" {
		t.Fatalf("source_revision = %q, want immutable v0.24.0 source commit", manifest.SourceRevision)
	}
	if manifest.MycelAPILine != "0.24" {
		t.Fatalf("mycel_api_line = %q, want 0.24", manifest.MycelAPILine)
	}
	if !reflect.DeepEqual(manifest.SupportedDaemonAPILines, []string{"0.24"}) {
		t.Fatalf("supported_daemon_api_lines = %v, want [0.24]", manifest.SupportedDaemonAPILines)
	}

	wantIncompatible := []incompatibleDaemonAPILine{
		{
			Line:      "0.20",
			Contracts: []string{"search-channel-owned-query-v1", "semantic-trigger-events-enum-v1"},
		},
		{
			Line:      "0.22",
			Contracts: []string{"semantic-projection-consistency-v1"},
		},
		{
			Line:      "0.23",
			Contracts: []string{"repository-backed-backup-restore-v1", "usage-accounting-admission-v1"},
		},
	}
	if !reflect.DeepEqual(manifest.IncompatibleDaemonAPILines, wantIncompatible) {
		t.Fatalf("incompatible_daemon_api_lines = %#v, want %#v", manifest.IncompatibleDaemonAPILines, wantIncompatible)
	}
	if manifest.PublishedAt != "2026-11-23T20:30:44Z" {
		t.Fatalf("published_at = %q, want v0.24.0 publication time", manifest.PublishedAt)
	}
}
