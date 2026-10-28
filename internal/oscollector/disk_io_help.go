//go:build linux || windows

package oscollector

// Help texts shared by the Linux and Windows disk I/O collectors, which
// expose the same series.
const (
	readTimeHelp    = "Total time spent on disk reads in seconds, summed over concurrent reads; rate() is the average read queue length"
	writeTimeHelp   = "Total time spent on disk writes in seconds, summed over concurrent writes; rate() is the average write queue length"
	queueLengthHelp = "Disk requests in progress at the time of the scrape"
)
