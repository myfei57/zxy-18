package aggregate

import (
	"fmt"
	"os"
	"path/filepath"

	"traceflow/internal/store"
)

// Window is one aggregation time bucket.
type Window struct {
	ID        string `json:"id"`
	Namespace string `json:"namespace"`
	Start     int64  `json:"start"`
	End       int64  `json:"end"`
	Closed    bool   `json:"closed"`
	SummaryID string `json:"summary_id,omitempty"`
}

// openWindow registers a window and persists it.
func (a *Aggregator) openWindow(namespace string, start, end int64) (Window, error) {
	w := Window{ID: windowID(namespace, start), Namespace: namespace, Start: start, End: end}
	if err := store.WriteJSON(a.windowPath(w.ID), w); err != nil {
		return Window{}, err
	}
	return w, nil
}

func windowID(namespace string, start int64) string {
	return fmt.Sprintf("%s-%d", namespace, start)
}

func (a *Aggregator) windowPath(id string) string {
	return filepath.Join(a.paths.WindowsDir(), id+".window.json")
}

// currentWindow returns the newest open window for a namespace.
func (a *Aggregator) currentWindow(namespace string) (Window, bool) {
	entries, err := os.ReadDir(a.paths.WindowsDir())
	if err != nil {
		return Window{}, false
	}
	var best Window
	found := false
	for _, entry := range entries {
		var w Window
		if err := store.ReadJSON(filepath.Join(a.paths.WindowsDir(), entry.Name()), &w); err != nil {
			continue
		}
		if w.Namespace != namespace || w.Closed {
			continue
		}
		if !found || w.End > best.End {
			best = w
			found = true
		}
	}
	return best, found
}

// EnsureCurrentWindow opens the current window for a namespace when missing.
func (a *Aggregator) EnsureCurrentWindow(namespace string, start, end int64) (Window, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if existing, ok := a.currentWindow(namespace); ok {
		return existing, nil
	}
	return a.openWindow(namespace, start, end)
}

// OpenWindows returns all open windows.
func (a *Aggregator) OpenWindows() []Window {
	entries, err := os.ReadDir(a.paths.WindowsDir())
	if err != nil {
		return nil
	}
	var out []Window
	for _, entry := range entries {
		var w Window
		if err := store.ReadJSON(filepath.Join(a.paths.WindowsDir(), entry.Name()), &w); err != nil {
			continue
		}
		if !w.Closed {
			out = append(out, w)
		}
	}
	return out
}

// CloseWindow durably writes the summary and only then marks the window closed.
func (a *Aggregator) CloseWindow(namespace, id string) (Summary, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	var w Window
	if err := store.ReadJSON(a.windowPath(id), &w); err != nil {
		return Summary{}, err
	}
	if w.Namespace != namespace {
		return Summary{}, fmt.Errorf("window %s belongs to namespace %s", id, w.Namespace)
	}
	if w.Closed {
		summary, _ := a.ReadSummary(w.SummaryID)
		return summary, nil
	}
	summary := BuildSummary(w, a.windowSpansLocked(w))
	if err := a.writeSummary(summary); err != nil {
		return Summary{}, err
	}
	w.Closed = true
	w.SummaryID = summary.ID
	if err := store.WriteJSON(a.windowPath(id), w); err != nil {
		return Summary{}, err
	}
	return summary, nil
}
