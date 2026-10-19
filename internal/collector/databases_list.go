package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// listAccessibleDatabases returns the online user databases the login can
// access (HAS_DBACCESS skips databases without a user for the login,
// mirrors, restoring and non-readable AG secondaries), optionally restricted
// to the given names (matched case-insensitively). Used by the collectors
// that read each database's own catalog views.
func listAccessibleDatabases(ctx context.Context, db *sql.DB, only []string) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT name
		FROM sys.databases
		WHERE database_id > 4
			AND state = 0
			AND HAS_DBACCESS(name) = 1
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}
	defer func() { _ = rows.Close() }()

	wanted := make(map[string]bool, len(only))
	for _, name := range only {
		wanted[strings.ToLower(name)] = true
	}

	var names []string
	var scanErrs []error
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan database row: %w", err))
			continue
		}
		if len(wanted) == 0 || wanted[strings.ToLower(name)] {
			names = append(names, name)
		}
	}
	return names, errors.Join(append(scanErrs, rows.Err())...)
}
