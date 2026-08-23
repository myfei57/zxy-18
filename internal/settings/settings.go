// Package settings defines TraceFlow runtime options and the on-disk layout.
package settings

// Options configures one TraceFlow instance.
type Options struct {
	// DataDir is the root of all durable state.
	DataDir string
	// Addr is the HTTP listen address.
	Addr string
	// Seed toggles demo data injection at bootstrap.
	Seed bool
}

// Defaults returns the default standalone options.
func Defaults() Options {
	return Options{DataDir: "data", Addr: "127.0.0.1:7790", Seed: true}
}
