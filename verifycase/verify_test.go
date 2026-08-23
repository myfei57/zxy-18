package verifycase

import (
	"os"
	"path/filepath"
	"testing"

	"traceflow/internal/aggregate"
	"traceflow/internal/audit"
	"traceflow/internal/ns"
	"traceflow/internal/quota"
	"traceflow/internal/settings"
	"traceflow/internal/span"
)

func TestAggregateWindowAfterSummaryDurable(t *testing.T) {
	root := filepath.Join(t.TempDir(), "data")
	paths := settings.NewPaths(root)
	for _, dir := range []string{root, paths.SpansDir(), paths.WindowsDir(), filepath.Dir(paths.CursorFile("aggregate"))} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	registry := ns.NewRegistry(filepath.Join(root, "registry.json"))
	if err := registry.Register(ns.New("ns-01", "svc")); err != nil {
		t.Fatal(err)
	}
	ledger := quota.NewLedger(paths.QuotaFile())
	st, err := span.NewStore(paths, ledger, registry)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	start := int64(1000)
	for _, spanID := range []string{"root-09", "child-09"} {
		parent := ""
		if spanID != "root-09" {
			parent = "root-09"
		}
		sp := span.New("trace-09", spanID, parent, "svc-i", "op-i")
		sp.Namespace = "ns-01"
		sp.StartedAt = start + 10
		if _, err := st.Append(sp); err != nil {
			t.Fatal(err)
		}
	}
	auditLogger := audit.NewLogger(paths.AuditFile())
	aggregator, err := aggregate.NewAggregator(paths, st, auditLogger)
	if err != nil {
		t.Fatal(err)
	}
	window, err := aggregator.EnsureCurrentWindow("ns-01", start, start+100)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aggregator.CloseWindow("ns-01", window.ID); err == nil {
		t.Fatal("closing must fail when the summary directory is missing")
	}
	found := false
	for _, open := range aggregator.OpenWindows() {
		if open.ID == window.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("window must remain open when its summary is not durable")
	}
}
