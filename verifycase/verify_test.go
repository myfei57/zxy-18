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

func TestReportCursorAfterAckDurable(t *testing.T) {
	root := filepath.Join(t.TempDir(), "data")
	paths := settings.NewPaths(root)
	for _, dir := range []string{
		root,
		paths.SpansDir(),
		paths.BatchesDir(),
		filepath.Dir(paths.CursorFile("report")),
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
	sp := span.New("trace-07", "span-07", "", "svc-g", "op-g")
	sp.Namespace = "ns-01"
	if _, err := st.Append(sp); err != nil {
		t.Fatal(err)
	}
	auditLogger := audit.NewLogger(paths.AuditFile())
	reporter, err := report.NewReporter(paths, st, report.NewLocalSink(filepath.Join(root, "sink")), auditLogger)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := reporter.Send("ns-01"); err == nil {
		t.Fatal("send must fail when the acknowledgement cannot be persisted")
	}
	if got := reporter.Cursor("ns-01"); got != 0 {
		t.Fatalf("cursor must not advance when the ack is not durable, got %d", got)
	}
}
