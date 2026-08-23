package verifycase

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"traceflow/internal/ns"
	"traceflow/internal/quota"
	"traceflow/internal/sample"
	"traceflow/internal/settings"
	"traceflow/internal/span"
)

func TestSampleRateSwitchRebuildsBudget(t *testing.T) {
	root := filepath.Join(t.TempDir(), "data")
	paths := settings.NewPaths(root)
	for _, dir := range []string{root, paths.SpansDir()} {
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
	sampler := sample.NewSampler(paths.SampleFile())
	if err := sampler.Switch(50, 100); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		sp := span.New(fmt.Sprintf("trace-%d", i), fmt.Sprintf("span-%d", i), "", "svc-h", "op-h")
		sp.Namespace = "ns-01"
		if _, err := sampler.Vote(st, "ns-01", sp); err != nil {
			t.Fatal(err)
		}
	}
	if err := sampler.Switch(5, 100); err != nil {
		t.Fatal(err)
	}
	next := span.New("trace-next", "span-next", "", "svc-h", "op-h")
	next.Namespace = "ns-01"
	decided, err := sampler.Vote(st, "ns-01", next)
	if err != nil {
		t.Fatal(err)
	}
	if !decided {
		t.Fatal("the rebuilt budget must apply to the next vote after a rate switch")
	}
}
