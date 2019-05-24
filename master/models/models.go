package models

import (
	"database/sql"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// contextKey is an unexported type for keys defined in this package for use
// with contexts.
type contextKey int

// ErrNoEntry is returned when a requested entry does not exist.
type ErrNoEntry struct {
	Type       string      // the type of the entry
	Identifier interface{} // identifier used to find the entry, usually unique
}

// IsErrNoEntry returns true if the error is an ErrNoEntry.
func IsErrNoEntry(err error) bool {
	_, ok := err.(*ErrNoEntry)
	return ok
}

// Error implements the error interface for ErrNoEntry.
func (err *ErrNoEntry) Error() string {
	return fmt.Sprintf("models: no entry for %s '%s'", err.Type, err.Identifier)
}

// ErrEmpty is returned when a table is empty.
type ErrEmpty struct {
	Name string // the name of the table
}

// IsErrEmpty returns true if the error is an ErrEmpty.
func IsErrEmpty(err error) bool {
	_, ok := err.(*ErrEmpty)
	return ok
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

// IsErrBadEffect returns true if the error is an ErrBadEffect.
func IsErrBadEffect(err error) bool {
	_, ok := err.(*ErrBadEffect)
	return ok
}

// Error implements the error interface for ErrBadEffect.
func (err *ErrBadEffect) Error() string {
	return fmt.Sprintf("%s: %d rows were affected, expected %d to be affected", err.Name,
		err.Affected, err.Expected)
}

// ErrInvalid is returned when a field of some model is invalid.
type ErrInvalid struct {
	Model string
	Which string
	Value string
}

// IsErrInvalid returns true if the error is an ErrInvalid.
func IsErrInvalid(err error) bool {
	_, ok := err.(*ErrInvalid)
	return ok
}

// Error implements the error interface for ErrInvalid.
func (err *ErrInvalid) Error() string {
	if err.Value == "" {
		return fmt.Sprintf("%s: %s cannot be blank", err.Model, err.Which)
	}
	return fmt.Sprintf("%s: invalid %s '%s'", err.Model, err.Which, err.Value)
}

// ShouldAffect takes an sql.Result and returns an error if the number of rows
// affected is different from what was expected.
func ShouldAffect(name string, res sql.Result, expected int64) error {
	if affected, err := res.RowsAffected(); err != nil {
		log.Panicf("%s: got error while fetching affected row count: %s", name, err)
	} else if affected != 1 {
		return &ErrBadEffect{Name: name, Affected: affected, Expected: expected}
	}

	return nil
}
