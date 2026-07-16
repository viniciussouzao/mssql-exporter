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

// CurrentOptions returns the current database connection options, falling
// back to the defaults when Connect has not been called yet.
func CurrentOptions() *Options {
	if actualOpts == nil {
		return DefaultOptions()
	}
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

// poolConfigurer is the subset of *sql.DB's pool-tuning methods Connect
// needs; it exists so applyPoolOptions can be tested without a real
// database connection.
type poolConfigurer interface {
	SetMaxOpenConns(int)
	SetMaxIdleConns(int)
	SetConnMaxLifetime(time.Duration)
	SetConnMaxIdleTime(time.Duration)
}

// applyPoolOptions pushes opts onto configurer. Zero lifetime/idle-time
// values are left at the driver default (a limit of 0 means "no limit" in
// database/sql, not "apply immediately").
func applyPoolOptions(configurer poolConfigurer, opts *Options) {
	configurer.SetMaxOpenConns(opts.MaxOpenConns)
	configurer.SetMaxIdleConns(opts.MaxIdleConns)

	if opts.ConnMaxLifetime > 0 {
		configurer.SetConnMaxLifetime(opts.ConnMaxLifetime)
	}

	if opts.ConnMaxIdleTime > 0 {
		configurer.SetConnMaxIdleTime(opts.ConnMaxIdleTime)
	}
}
