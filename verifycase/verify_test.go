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

func TestSpanFinishAfterPayloadDurable(t *testing.T) {
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
	st, err := span.NewStore(paths, ledger, registry)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	sp := span.New("trace-04", "span-04", "", "svc-d", "op-d")
	sp.Namespace = "ns-01"
	if _, err := st.Append(sp); err != nil {
		t.Fatal(err)
	}
	if err := st.Finish("ns-01", "span-04", map[string]string{"k": "v"}); err == nil {
		t.Fatal("finish must fail when the payload directory is missing")
	}
	if err := st.Load(registry); err != nil {
		t.Fatal(err)
	}
	for _, existing := range st.List("ns-01") {
		if existing.Finished() {
			t.Fatal("span must not be marked finished when its payload is not durable")
		}
	}
}
