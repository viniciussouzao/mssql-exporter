package collector

import "time"

// Options tunes collectors that need more than the on/off toggle.
type Options struct {
	TopQueries         TopQueriesOptions
	IndexFragmentation IndexFragmentationOptions
	QueryStore         QueryStoreOptions
}

// TopQueriesOptions configures the top_queries collector.
type TopQueriesOptions struct {
	// Limit is the number of queries (by total CPU time) exposed per scrape.
	Limit int
	// IncludeText adds sqlserver_top_query_info with the statement text,
	// literals obfuscated and truncated to TextLength characters.
	IncludeText bool
	TextLength  int
}

// IndexFragmentationOptions configures the index_fragmentation collector,
// which collects in the background instead of on every scrape.
type IndexFragmentationOptions struct {
	// Interval between background collections; each one is bounded by it.
	Interval time.Duration
	// MinPageCount skips indexes smaller than this (8 KB pages): their
	// fragmentation is meaningless and they dominate the series count.
	MinPageCount int
	// Databases restricts the collection to these databases; empty means
	// every online user database the login can access.
	Databases []string
}

// QueryStoreOptions configures the query_store collector, which collects in
// the background instead of on every scrape.
type QueryStoreOptions struct {
	// Interval between background collections; each one is bounded by it.
	Interval time.Duration
	// Databases restricts the collection to these databases; empty means
	// every online user database the login can access.
	Databases []string
}

// Option bounds, enforced by the config package.
const (
	TopQueriesMaxLimit      = 1000
	TopQueriesMinTextLength = 16
	TopQueriesMaxTextLength = 4000

	IndexFragmentationMinInterval = time.Minute
	QueryStoreMinInterval         = time.Minute
)

// DefaultOptions returns the options used when none are configured.
func DefaultOptions() Options {
	return Options{
		TopQueries: TopQueriesOptions{
			Limit:      50,
			TextLength: 120,
		},
		IndexFragmentation: IndexFragmentationOptions{
			Interval:     6 * time.Hour,
			MinPageCount: 1000,
		},
		QueryStore: QueryStoreOptions{
			Interval: 5 * time.Minute,
		},
	}
}
