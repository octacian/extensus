package models

import (
	"database/sql"
	"fmt"
	"github.com/octacian/extensus/core"
	"golang.org/x/crypto/bcrypt"
	"time"
)

// User identifies an account.
type User struct {
	ID       uint64
	Created  time.Time
	Modified time.Time

	Name     string
	Email    string
	Password []byte
}

// NewUser takes a name, email, and plaintext password and returns a new User.
// If an error occurs while hashing the password, it is returned.
func NewUser(name, email, password string) (*User, error) {
	user := &User{
		Created:  core.Time(),
		Modified: core.Time(),
		Name:     name,
		Email:    email,
	}

	if err := user.SetPassword(password); err != nil {
		return nil, err
	}

	return user, nil
}

// ListUser returns an array of all Users in the database. If the user table is
// empty an ErrEmpty is returned. If anything else goes wrong it is returned.
func ListUser() ([]User, error) {
	users := []User{}
	err := core.GetDB().Select(&users, "SELECT * FROM user")
	if len(users) == 0 {
		return nil, &ErrEmpty{"user"}
	}

	return users, err
}

// GetUser fetches a User from the database by email. If no such user exists
// an error is returned.
func GetUser(email string) (*User, error) {
	user := &User{}
	row := core.GetDB().QueryRowx("SELECT * FROM user WHERE Email=?", email)

	if err := row.StructScan(user); err == sql.ErrNoRows {
		return nil, &ErrNoEntry{Type: "user", Identifier: email}
	} else if err != nil {
		return nil, err
	}

	return user, nil
}

// Save propagates any changes back to the database. If the ID field is 0, a
// new entry is created. Otherwise, Save attempts to update an existing entry.
// If anything goes wrong an error is returned.
func (user *User) Save() error {
	if user.ID == 0 {
		res, err := core.GetDB().Exec("INSERT INTO user (Created, Modified, Name, Email, Password) VALUES (?, ?, ?, ?, ?)",
			user.Created, user.Modified, user.Name, user.Email, user.Password)
		if err != nil {
			return err
		}

		if insertID, err := res.LastInsertId(); err != nil {
			panic(fmt.Sprint("User.Save: got error while fetching ID of inserted user:\n", err))
		} else {
			user.ID = uint64(insertID)
		}
	} else {
		user.Modified = core.Time()
		res, err := core.GetDB().Exec("UPDATE user SET Modified=?, Name=?, Email=?, Password=? WHERE Email=?",
			user.Modified, user.Name, user.Email, user.Password, user.Email)
		if err != nil {
			return err
		}

		return ShouldAffect("User.Save", res, 1)
	}

	return nil
}

// Delete removes the user from the database. If the ID field is 0 an
// ErrNoEntry is returned. If any other errors occurs it is returned.
func (user *User) Delete() error {
	res, err := core.GetDB().Exec("DELETE FROM user WHERE Email=?", user.Email)
	if err != nil {
		return err
	}

	return ShouldAffect("User.Delete", res, 1)
}

// SetPassword takes a plaintext password and hashes it before storing it in
// the password field. If an error occurs while hashing the password, it is
// returned.
func (user *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), core.GetConfig().HashCost)
	if err != nil {
		return err
	}

	user.Password = hash
	return nil
}

// Authenticate takes a plaintext password and compares it with the hashed
// password stored. Returns nil on succcess or an error on failure.
func (user *User) Authenticate(password string) error {
	return bcrypt.CompareHashAndPassword(user.Password, []byte(password))
}
