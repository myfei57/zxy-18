package verifycase

import (
	"os"
	"path/filepath"
	"testing"

	"traceflow/internal/ns"
	"traceflow/internal/quota"
	"traceflow/internal/sample"
	"traceflow/internal/settings"
	"traceflow/internal/span"
)

func TestSampleAfterSpanDurable(t *testing.T) {
	root := filepath.Join(t.TempDir(), "data")
	paths := settings.NewPaths(root)
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
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
	sampler := sample.NewSampler(paths.SampleFile())
	if err := sampler.Switch(100, 100); err != nil {
		t.Fatal(err)
	}
	sp := span.New("trace-02", "span-02", "", "svc-b", "op-b")
	sp.Namespace = "ns-01"
	if _, err := sampler.Vote(st, "ns-01", sp); err == nil {
		t.Fatal("vote must fail when the span journal cannot be written")
	}
	reloaded := sample.NewSampler(paths.SampleFile())
	if err := reloaded.Load(); err != nil {
		t.Fatal(err)
	}
	if reloaded.Sampled != 0 {
		t.Fatalf("sampled marker must not advance when span data is not durable, got %d", reloaded.Sampled)
	}
}
