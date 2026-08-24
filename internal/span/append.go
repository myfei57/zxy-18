package span

import "encoding/json"

// appendRecord writes one record to the namespace journal and makes it durable.
func (st *Store) appendRecordLocked(namespace string, rec journalRecord) (uint64, error) {
	seq := st.nextSeqLocked(namespace)
	payload, err := json.Marshal(rec)
	if err != nil {
		return 0, err
	}
	journal, err := st.journalLocked(namespace)
	if err != nil {
		return 0, err
	}
	if err := journal.Append(seq, payload); err != nil {
		return 0, err
	}
	if err := journal.Commit(seq); err != nil {
		return 0, err
	}
	return seq, nil
}

// Append validates, quota-gates and durably stores a span record.
func (st *Store) Append(s Span) (uint64, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}
	st.mu.Lock()
	defer st.mu.Unlock()
	s.Seq = st.nextSeqLocked(s.Namespace)
	seq, err := st.appendRecordLocked(s.Namespace, journalRecord{Kind: "span", Span: s})
	if err != nil {
		return 0, err
	}
	if err := st.quota.Check(s.Namespace, 1); err != nil {
		return 0, err
	}
	st.spans[s.Namespace] = append(st.spans[s.Namespace], s)
	if err := st.quota.Occupy(s.Namespace, 1); err != nil {
		return 0, err
	}
	return seq, nil
}
