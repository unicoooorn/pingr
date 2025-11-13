package infographics

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/unicoooorn/pingr/internal/config"
	"github.com/unicoooorn/pingr/internal/model"
)

func TestBuildDOTFromConfig_Basic(t *testing.T) {
	cfg := config.Config{
		Backends: map[string]config.BackendConfig{
			"A": {Deps: []string{"B"}},
			"B": {Deps: []string{}},
		},
	}

	infos := map[string]model.SubsystemInfo{
		"A": {Check: model.CheckResult{Status: model.PingStatusOk}},
		"B": {Check: model.CheckResult{Status: model.PingStatusNotOk}},
	}

	ir := NewImageRenderer(cfg, 0)
	dot := ir.buildDOTFromConfig(infos)

	if !strings.Contains(dot, `"A" [label=<<TABLE`) {
		t.Fatalf("dot missing HTML label for node A: %s", dot)
	}
	if !strings.Contains(dot, `"B" [label=<<TABLE`) {
		t.Fatalf("dot missing HTML label for node B: %s", dot)
	}
	
	if !strings.Contains(dot, `fillcolor="#7ed07e"`) {
		t.Fatalf("expected #7ed07e in dot for ok node; got: %s", dot)
	}
	if !strings.Contains(dot, `fillcolor="#ff6b6b"`) {
		t.Fatalf("expected #ff6b6b in dot for not ok node; got: %s", dot)
	}
	
	if !strings.Contains(dot, `"A" -> "B"`) {
		t.Fatalf("expected edge A -> B in dot; got: %s", dot)
	}
	
	if !strings.Contains(dot, "digraph G") {
		t.Fatalf("expected digraph G in dot; got: %s", dot)
	}
	if !strings.Contains(dot, "rankdir=TB") {
		t.Fatalf("expected rankdir=TB in dot; got: %s", dot)
	}
}

func TestRender_ReturnsPNG(t *testing.T) {
	cfg := config.Config{
		Backends: map[string]config.BackendConfig{
			"A": {Deps: []string{"B"}},
			"B": {Deps: []string{}},
		},
	}

	infos := map[string]model.SubsystemInfo{
		"A": {Check: model.CheckResult{Status: model.PingStatusOk}},
		"B": {Check: model.CheckResult{Status: model.PingStatusOk}},
	}

	ir := NewImageRenderer(cfg, 5*time.Second)
	ctx := context.Background()

	img, err := ir.Render(ctx, infos)
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	if len(img) == 0 {
		t.Fatalf("Render returned empty image")
	}

	if !bytes.HasPrefix(img, []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}) {
		t.Fatalf("rendered image does not look like PNG (bad header): %x...", img[:8])
	}
}

func TestComputeDepths(t *testing.T) {
	cfg := config.Config{
		Backends: map[string]config.BackendConfig{
			"A": {Deps: []string{"B", "C"}},
			"B": {Deps: []string{"C"}},
			"C": {Deps: []string{}},
			"D": {Deps: []string{}},
		},
	}

	depths, maxDepth := computeDepths(cfg)
	
	expectedDepths := map[string]int{
		"A": 2,
		"B": 1,
		"C": 0,
		"D": 0,
	}
	
	if maxDepth != 2 {
		t.Errorf("expected maxDepth=2, got %d", maxDepth)
	}
	
	for node, expectedDepth := range expectedDepths {
		if depths[node] != expectedDepth {
			t.Errorf("node %s: expected depth %d, got %d", node, expectedDepth, depths[node])
		}
	}
}

func TestStatusToColor(t *testing.T) {
	tests := []struct {
		status model.PingStatus
		want   string
	}{
		{model.PingStatusOk, "#7ed07e"},
		{model.PingStatusNotOk, "#ff6b6b"},
	}
	
	for _, tt := range tests {
		got := statusToColor(tt.status)
		if got != tt.want {
			t.Errorf("statusToColor(%v) = %v, want %v", tt.status, got, tt.want)
		}
	}
}

func TestEscapeID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"simple", `"simple"`},
		{"with space", `"with space"`},
		{"with-dash", `"with-dash"`},
		{"with_underscore", `"with_underscore"`},
	}
	
	for _, tt := range tests {
		got := escapeID(tt.input)
		if got != tt.want {
			t.Errorf("escapeID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestBuildDOTFromConfig_EmptyDeps(t *testing.T) {
	cfg := config.Config{
		Backends: map[string]config.BackendConfig{
			"A": {Deps: []string{}},
			"B": {Deps: []string{""}},
		},
	}

	infos := map[string]model.SubsystemInfo{
		"A": {Check: model.CheckResult{Status: model.PingStatusOk}},
		"B": {Check: model.CheckResult{Status: model.PingStatusOk}},
	}

	ir := NewImageRenderer(cfg, 0)
	dot := ir.buildDOTFromConfig(infos)

	if strings.Contains(dot, `"B" -> ""`) {
		t.Fatalf("dot should not contain edge from B to empty string: %s", dot)
	}
	
	if !strings.Contains(dot, `"A" [label=<<TABLE`) {
		t.Fatalf("dot missing node A: %s", dot)
	}
	if !strings.Contains(dot, `"B" [label=<<TABLE`) {
		t.Fatalf("dot missing node B: %s", dot)
	}
}

func TestHTMLEscape(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"normal", "normal"},
		{"a & b", "a &amp; b"},
		{"a < b", "a &lt; b"},
		{"a > b", "a &gt; b"},
	}
	
	for _, tt := range tests {
		got := htmlEscape(tt.input)
		if got != tt.want {
			t.Errorf("htmlEscape(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}