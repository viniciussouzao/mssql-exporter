package db

import (
	"context"
	"database/sql"
)

var db *sql.DB

// Connect initializes the database connection with the given connection string and options.
func Connect(connString string, opts *Options) error {
	var err error

	actualOpts = opts

	db, err = sql.Open("sqlserver", connString)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(opts.MaxOpenConns)
	db.SetMaxIdleConns(opts.MaxIdleConns)

	if opts.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(opts.ConnMaxLifetime)
	}

	if opts.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(opts.ConnMaxIdleTime)
	}

	ctx, cancel := context.WithTimeout(context.Background(), opts.ConnTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

// DB returns the database connection instance.
func DB() *sql.DB {
	return db
}

// // Timeout returns the timeout duration for database operations.
// func Timeout() time.Duration {
// 	return opts.ConnTimeout
// }
