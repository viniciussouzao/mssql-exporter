package db

import (
	"time"
)

// Options holds the database connection configuration options.
type Options struct {
	//MaxOpenConns sets the maximum number of open connections to the database(in use + idle).
	MaxOpenConns int

	//MaxIdleConns sets the maximum number of connections in the idle connection pool.
	MaxIdleConns int

	//ConnMaxLifetime sets the maximum amount of time a connection may be reused.
	ConnMaxLifetime time.Duration

	//ConnMaxIdleTime sets the maximum amount of time a connection may be idle.
	ConnMaxIdleTime time.Duration

	//ConnTimeout sets the timeout for queries.
	ConnTimeout time.Duration
}

var actualOpts *Options

// CurrentOptions returns the current database connection options.
func CurrentOptions() *Options {
	return actualOpts
}

// DefaultOptions returns the default database connection options.
func DefaultOptions() *Options {
	return &Options{
		MaxOpenConns:    5,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 2 * time.Minute,
		ConnTimeout:     5 * time.Second,
	}
}
