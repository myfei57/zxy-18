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

func TestAggregateCursorAfterResultDurable(t *testing.T) {
	root := filepath.Join(t.TempDir(), "data")
	paths := settings.NewPaths(root)
	for _, dir := range []string{root, paths.SpansDir(), filepath.Dir(paths.CursorFile("aggregate"))} {
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
	for _, spanID := range []string{"root-03", "child-03"} {
		parent := ""
		if spanID != "root-03" {
			parent = "root-03"
		}
		sp := span.New("trace-03", spanID, parent, "svc-c", "op-c")
		sp.Namespace = "ns-01"
		if _, err := st.Append(sp); err != nil {
			t.Fatal(err)
		}
	}
	auditLogger := audit.NewLogger(paths.AuditFile())
	aggregator, err := aggregate.NewAggregator(paths, st, auditLogger)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aggregator.Run("ns-01", "trace-03"); err == nil {
		t.Fatal("aggregation must fail when the result directory is missing")
	}
	if got := aggregator.CursorPosition("ns-01"); got != 0 {
		t.Fatalf("aggregate cursor must stay at 0 when the result is not durable, got %d", got)
	}
}
