package verifycase

import (
	"os"
	"path/filepath"
	"testing"

	"traceflow/internal/ns"
	"traceflow/internal/quota"
	"traceflow/internal/settings"
	"traceflow/internal/span"
)

func TestSpanQuotaRejectsBeforeWrite(t *testing.T) {
	root := filepath.Join(t.TempDir(), "data")
	paths := settings.NewPaths(root)
	if err := os.MkdirAll(paths.SpansDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	registry := ns.NewRegistry(filepath.Join(root, "registry.json"))
	if err := registry.Register(ns.New("ns-01", "svc")); err != nil {
		t.Fatal(err)
	}
	ledger := quota.NewLedger(paths.QuotaFile())
	if err := ledger.SetLimit("ns-01", 1); err != nil {
		t.Fatal(err)
	}
	if err := ledger.Occupy("ns-01", 1); err != nil {
		t.Fatal(err)
	}
	st, err := span.NewStore(paths, ledger, registry)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	sp := span.New("trace-05", "span-05", "", "svc-e", "op-e")
	sp.Namespace = "ns-01"
	if _, err := st.Append(sp); err == nil {
		t.Fatal("append must reject an over-quota span")
	}
	if err := st.Load(registry); err != nil {
		t.Fatal(err)
	}
	if got := st.Count("ns-01"); got != 0 {
		t.Fatalf("over-quota span must not be stored, got %d spans", got)
	}
}
