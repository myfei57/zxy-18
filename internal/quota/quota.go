// Package quota enforces per-namespace span capacity.
package quota

// DefaultLimit is the per-namespace span ceiling.
const DefaultLimit = 10000

// Status is a namespace quota snapshot.
type Status struct {
	Namespace string `json:"namespace"`
	Limit     int    `json:"limit"`
	Used      int    `json:"used"`
	Remaining int    `json:"remaining"`
}
