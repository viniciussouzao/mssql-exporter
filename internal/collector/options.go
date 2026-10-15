package collector

// Options tunes collectors that need more than the on/off toggle.
type Options struct {
	TopQueries TopQueriesOptions
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

// Option bounds, enforced by the config package.
const (
	TopQueriesMaxLimit      = 1000
	TopQueriesMinTextLength = 16
	TopQueriesMaxTextLength = 4000
)

// DefaultOptions returns the options used when none are configured.
func DefaultOptions() Options {
	return Options{
		TopQueries: TopQueriesOptions{
			Limit:      50,
			TextLength: 120,
		},
	}
}
