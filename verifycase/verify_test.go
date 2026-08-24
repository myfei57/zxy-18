package verifycase

import (
	"os"
	"path/filepath"
	"testing"

	"traceflow/internal/index"
	"traceflow/internal/settings"
	"traceflow/internal/trace"
)

func TestTraceQueryUsesCurrentIndex(t *testing.T) {
	root := filepath.Join(t.TempDir(), "data")
	paths := settings.NewPaths(root)
	if err := os.MkdirAll(paths.IndexDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	idx := index.NewIndex(paths.IndexDir())
	if _, err := idx.Build([]index.BuildInput{{
		TraceID: "trace-a", Namespace: "ns-01", RootSpanID: "r-a", SpanCount: 1, State: "aggregated",
	}}); err != nil {
		t.Fatal(err)
	}
	resolver := trace.NewResolver(idx)
	if _, ok := resolver.Query("trace-a"); !ok {
		t.Fatal("the first generation must resolve a query")
	}
	if _, err := idx.Build([]index.BuildInput{{
		TraceID: "trace-b", Namespace: "ns-01", RootSpanID: "r-b", SpanCount: 1, State: "aggregated",
	}}); err != nil {
		t.Fatal(err)
	}
	if _, ok := resolver.Query("trace-b"); !ok {
		t.Fatal("query must read the current index generation")
	}
}
