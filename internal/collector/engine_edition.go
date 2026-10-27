package collector

import (
	"context"
	"database/sql"
	"fmt"
)

// engineEditionManagedInstance is SERVERPROPERTY('EngineEdition') on Azure SQL
// Managed Instance.
const engineEditionManagedInstance = 8

// queryEngineEdition returns SERVERPROPERTY('EngineEdition'), used by the
// collectors whose source differs on Azure SQL Managed Instance.
func queryEngineEdition(ctx context.Context, db *sql.DB) (int, error) {
	var engineEdition int
	if err := db.QueryRowContext(ctx,
		`SELECT CAST(SERVERPROPERTY('EngineEdition') AS int)`,
	).Scan(&engineEdition); err != nil {
		return 0, fmt.Errorf("failed to query engine edition: %w", err)
	}
	return engineEdition, nil
}
