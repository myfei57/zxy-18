package trace

import (
	"fmt"
	"sort"

	"traceflow/internal/span"
)

// Node is one span in an assembled call tree.
type Node struct {
	Span     span.Span `json:"span"`
	Children []Node    `json:"children,omitempty"`
}

// Assemble builds a trace record and its call tree from stored spans.
func Assemble(namespace, traceID string, spans []span.Span) (Trace, []Node, error) {
	if len(spans) == 0 {
		return Trace{}, nil, fmt.Errorf("trace %s has no spans", traceID)
	}
	rootID := rootSpanID(spans)
	tr := Trace{
		TraceID:    traceID,
		Namespace:  namespace,
		RootSpanID: rootID,
		State:      StateReported,
		SpanCount:  len(spans),
	}
	return tr, buildTree(rootID, spans), nil
}

func rootSpanID(spans []span.Span) string {
	byID := map[string]span.Span{}
	for _, existing := range spans {
		byID[existing.SpanID] = existing
	}
	for _, existing := range spans {
		if existing.ParentID == "" {
			return existing.SpanID
		}
		if _, ok := byID[existing.ParentID]; !ok {
			return existing.SpanID
		}
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i].StartedAt < spans[j].StartedAt })
	return spans[0].SpanID
}

func buildTree(rootID string, spans []span.Span) []Node {
	byParent := map[string][]span.Span{}
	for _, existing := range spans {
		byParent[existing.ParentID] = append(byParent[existing.ParentID], existing)
	}
	nodes := make([]Node, 0, len(byParent[rootID]))
	for _, existing := range byParent[rootID] {
		nodes = append(nodes, buildNode(existing, byParent))
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Span.StartedAt < nodes[j].Span.StartedAt })
	return nodes
}

func buildNode(existing span.Span, byParent map[string][]span.Span) Node {
	node := Node{Span: existing}
	for _, child := range byParent[existing.SpanID] {
		node.Children = append(node.Children, buildNode(child, byParent))
	}
	return node
}

// Depth returns the depth of a call tree.
func Depth(nodes []Node) int {
	max := 0
	for _, node := range nodes {
		depth := 1 + Depth(node.Children)
		if depth > max {
			max = depth
		}
	}
	return max
}
