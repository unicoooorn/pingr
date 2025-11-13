package infographics

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	"os/exec"

	"github.com/unicoooorn/pingr/internal/config"
	"github.com/unicoooorn/pingr/internal/model"
	"github.com/unicoooorn/pingr/internal/service"
)

var _ service.InfographicsRenderer = &ImageRenderer{}

type ImageRenderer struct {
	cfg     config.Config
	timeout time.Duration
}

func NewImageRenderer(cfg config.Config, timeout time.Duration) *ImageRenderer {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &ImageRenderer{
		cfg:     cfg,
		timeout: timeout,
	}
}

func (ir *ImageRenderer) Render(ctx context.Context, infos map[string]model.SubsystemInfo) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	dot := ir.buildDOTFromConfig(infos)

	return renderDOTToPNG(ctx, dot, ir.timeout)
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func htmlLabelFor(labelText string, fontFace string, pointSize int) string {
	if labelText == "" {
		labelText = ""
	}
	labelText = strings.ReplaceAll(labelText, "\n", "<BR/>")
	esc := htmlEscape(labelText)
	return fmt.Sprintf("<<TABLE BORDER=\"0\" CELLBORDER=\"0\" CELLSPACING=\"0\"><TR><TD ALIGN=\"center\" VALIGN=\"middle\"><FONT FACE=\"%s\" POINT-SIZE=\"%d\">%s</FONT></TD></TR></TABLE>>",
		fontFace, pointSize, esc)
}

func (ir *ImageRenderer) buildDOTFromConfig(infos map[string]model.SubsystemInfo) string {
	depths, maxDepth := computeDepths(ir.cfg)

	var b strings.Builder
	b.WriteString("digraph G {\n")
	b.WriteString("rankdir=TB;\n")
	b.WriteString("bgcolor=\"white\";\n")
	b.WriteString("graph [layout=dot, dpi=200];\n")
	b.WriteString("splines=true; overlap=false; nodesep=1.0; ranksep=1.2;\n")

	const fontFace = "DejaVu Sans"
	const fontSize = 12
	const margin = 0.25
	b.WriteString(fmt.Sprintf(
		"node [shape=circle, style=filled, color=\"none\", fontname=\"%s\", fixedsize=true, fontsize=%d, margin=%g, penwidth=1.2];\n",
		fontFace, fontSize, margin))

	b.WriteString("edge [color=\"#333333\", penwidth=1.0, arrowsize=0.9, arrowhead=normal, headclip=true, tailclip=true];\n")

	ranks := make(map[int][]string)
	for name := range ir.cfg.Backends {
		depth := depths[name]
		rankIdx := maxDepth - depth
		ranks[rankIdx] = append(ranks[rankIdx], name)
	}

	for name, backend := range ir.cfg.Backends {
		labelText := name
		labelHTML := htmlLabelFor(labelText, fontFace, fontSize)

		status := model.PingStatus(infos[name].Check.Status)
		color := statusToColor(status)

		b.WriteString(fmt.Sprintf("%s [label=%s, fillcolor=%s];\n",
			escapeID(name), labelHTML, strconv.Quote(color)))

		for _, dep := range backend.Deps {
			if dep == "" {
				continue
			}
			b.WriteString(fmt.Sprintf("%s -> %s;\n", escapeID(name), escapeID(dep)))
		}
	}

	for rankIdx := 0; rankIdx <= maxDepth; rankIdx++ {
		nodes := ranks[rankIdx]
		if len(nodes) == 0 {
			continue
		}
		b.WriteString("subgraph {\n  rank = same;\n")
		for _, n := range nodes {
			b.WriteString("  " + escapeID(n) + ";\n")
		}
		b.WriteString("}\n")
	}

	b.WriteString("}\n")
	return b.String()
}

func renderDOTToPNG(ctx context.Context, dot string, timeout time.Duration) ([]byte, error) {
	ctx2, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if _, err := exec.LookPath("dot"); err != nil {
		return nil, fmt.Errorf("graphviz not installed: %w", err)
	}

	cmd := exec.CommandContext(ctx2, "dot", "-Tpng")
	cmd.Stdin = strings.NewReader(dot)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("execute dot command: %w, stderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

func statusToColor(status model.PingStatus) string {
	switch status {
	case model.PingStatusOk:
		return "#7ed07e"
	case model.PingStatusNotOk:
		return "#ff6b6b"
	default:
		return "#e6e6e6"
	}
}

func escapeID(id string) string {
	return strconv.Quote(id)
}

func computeDepths(cfg config.Config) (map[string]int, int) {
	depths := make(map[string]int)
	var dfs func(string) int
	dfs = func(name string) int {
		if d, ok := depths[name]; ok {
			return d
		}
		bc, ok := cfg.Backends[name]
		if !ok {
			depths[name] = 0
			return 0
		}
		if len(bc.Deps) == 0 {
			depths[name] = 0
			return 0
		}
		maxd := 0
		for _, dep := range bc.Deps {
			if dep == name {
				continue
			}
			dd := dfs(dep)
			if dd > maxd {
				maxd = dd
			}
		}
		depths[name] = maxd + 1
		return depths[name]
	}

	maxDepth := 0
	for name := range cfg.Backends {
		d := dfs(name)
		if d > maxDepth {
			maxDepth = d
		}
	}
	return depths, maxDepth
}