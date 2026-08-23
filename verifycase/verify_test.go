package verifycase

import (
	"os"
	"path/filepath"
	"testing"

	"traceflow/internal/audit"
	"traceflow/internal/ns"
	"traceflow/internal/quota"
	"traceflow/internal/report"
	"traceflow/internal/settings"
	"traceflow/internal/span"
)

func TestAuditAfterBatchDurable(t *testing.T) {
	root := filepath.Join(t.TempDir(), "data")
	paths := settings.NewPaths(root)
	for _, dir := range []string{
		root,
		paths.SpansDir(),
		filepath.Dir(paths.CursorFile("report")),
		paths.AckDir(),
	} {
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
	sp := span.New("trace-10", "span-10", "", "svc-j", "op-j")
	sp.Namespace = "ns-01"
	if _, err := st.Append(sp); err != nil {
		t.Fatal(err)
	}
	auditLogger := audit.NewLogger(paths.AuditFile())
	reporter, err := report.NewReporter(paths, st, report.NewLocalSink(filepath.Join(root, "sink")), auditLogger)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := reporter.Flush("ns-01"); err == nil {
		t.Fatal("flush must fail when the batch directory is missing")
	}
	for _, event := range auditLogger.Recent(50) {
		if event.Kind == audit.KindReport && event.Result == "ok" {
			t.Fatal("audit must not record a success before the batch is durable")
		}
	}
}
