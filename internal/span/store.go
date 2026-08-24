package span

import (
	"encoding/json"
	"fmt"
	"sync"

	"traceflow/internal/ns"
	"traceflow/internal/quota"
	"traceflow/internal/settings"
	"traceflow/internal/store"
)

// journalRecord is a typed entry in a namespace span journal.
type journalRecord struct {
	Kind string `json:"kind"`
	Span Span   `json:"span,omitempty"`
}

// Store persists span journals per namespace and exposes reads.
type Store struct {
	paths    settings.Paths
	quota    *quota.Ledger
	mu       sync.Mutex
	journals map[string]*store.Journal
	seqs     map[string]uint64
	spans    map[string][]Span
}

// NewStore opens a span store for every registered namespace.
func NewStore(paths settings.Paths, ledger *quota.Ledger, registry *ns.Registry) (*Store, error) {
	st := &Store{
		paths:    paths,
		quota:    ledger,
		journals: map[string]*store.Journal{},
		seqs:     map[string]uint64{},
		spans:    map[string][]Span{},
	}
	if err := st.Load(registry); err != nil {
		return nil, err
	}
	return st, nil
}

// Load recovers every namespace journal and rebuilds the in-memory view.
func (st *Store) Load(registry *ns.Registry) error {
	st.mu.Lock()
	defer st.mu.Unlock()
	for _, namespace := range registry.List() {
		records, err := store.Recover(
			ns.SpanJournalPath(st.paths, namespace),
			ns.SpanCommitPath(st.paths, namespace),
		)
		if err != nil {
			return err
		}
		var spans []Span
		for _, rec := range records {
			var jr journalRecord
			if err := json.Unmarshal(rec.Payload, &jr); err != nil {
				return fmt.Errorf("decode span record: %w", err)
			}
			if jr.Span.SpanID == "" {
				continue
			}
			spans = upsertSpan(spans, jr.Span)
			if jr.Span.Seq > st.seqs[namespace.Name] {
				st.seqs[namespace.Name] = jr.Span.Seq
			}
		}
		st.spans[namespace.Name] = spans
	}
	return nil
}

func upsertSpan(spans []Span, incoming Span) []Span {
	for i, existing := range spans {
		if existing.SpanID == incoming.SpanID {
			spans[i] = incoming
			return spans
		}
	}
	return append(spans, incoming)
}

// Close flushes and closes all open journals.
func (st *Store) Close() error {
	st.mu.Lock()
	defer st.mu.Unlock()
	for _, journal := range st.journals {
		if err := journal.Close(); err != nil {
			return err
		}
	}
	st.journals = map[string]*store.Journal{}
	return nil
}

func (st *Store) journalLocked(namespace string) (*store.Journal, error) {
	if journal, ok := st.journals[namespace]; ok {
		return journal, nil
	}
	journal, err := store.OpenJournal(
		ns.SpanJournalPath(st.paths, ns.Namespace{Name: namespace}),
		ns.SpanCommitPath(st.paths, ns.Namespace{Name: namespace}),
	)
	if err != nil {
		return nil, err
	}
	st.journals[namespace] = journal
	return journal, nil
}

func (st *Store) nextSeqLocked(namespace string) uint64 {
	st.seqs[namespace]++
	return st.seqs[namespace]
}

func (st *Store) spanByIDLocked(namespace, spanID string) (Span, bool) {
	for _, existing := range st.spans[namespace] {
		if existing.SpanID == spanID {
			return existing, true
		}
	}
	return Span{}, false
}
