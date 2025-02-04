package initialise

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgconn"
)

func exec(db *sql.DB, stmt string, possibleErrCodes []string, args ...interface{}) error {
	_, err := db.Exec(stmt, args...)
	pgErr := new(pgconn.PgError)
	if errors.As(err, &pgErr) {
		for _, possibleCode := range possibleErrCodes {
			if possibleCode == pgErr.Code {
				return nil
			}
		}
	}
	return err
}
