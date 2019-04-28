package models

import (
	"database/sql"
	"fmt"
)

// ErrNoEntry is returned when a requested entry does not exist.
type ErrNoEntry struct {
	Type       string // the type of the entry
	Identifier string // identifier used to find the entry, usually unique
}

// Error implements the error interface for ErrNoEntry.
func (err *ErrNoEntry) Error() string {
	return fmt.Sprintf("models: no entry for %s '%s'", err.Type, err.Identifier)
}

// ErrEmpty is returned when a table is empty.
type ErrEmpty struct {
	Name string // the name of the table
}

// Error implements the error interface for ErrEmpty.
func (err *ErrEmpty) Error() string {
	return fmt.Sprintf("models: %s table is empty", err.Name)
}

// ErrBadEffect is returned when some SQL affects an unexpected number of rows.
type ErrBadEffect struct {
	Name     string // a name to identify the caller
	Affected int64  // the number of rows that were affected
	Expected int64  // the number of rows expected to be affected
}

// Error implements the error interface for ErrBadEffect.
func (err *ErrBadEffect) Error() string {
	return fmt.Sprintf("%s: %d rows were affected, expected %d to be affected", err.Name,
		err.Affected, err.Expected)
}

// ShouldAffect takes an sql.Result and returns an error if the number of rows
// affected is different from what was expected.
func ShouldAffect(name string, res sql.Result, expected int64) error {
	if affected, err := res.RowsAffected(); err != nil {
		panic(fmt.Sprintf("%s: got error while fetching affected row count:\n%s", name, err))
	} else if affected != 1 {
		return &ErrBadEffect{Name: name, Affected: affected, Expected: expected}
	}

	return nil
}
