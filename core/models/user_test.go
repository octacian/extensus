package models

import (
	"testing"

	"github.com/octacian/extensus/core"
)

var testPassword = "!test?9@_*"

// WithUser executes a closure providing it with a valid, generic user.
func WithUser(t *testing.T, fn func(*User)) {
	if user, err := NewUser("John Doe", "john@doe.me", testPassword); err != nil {
		t.Error("NewUser: got error:\n", err)
	} else {
		fn(user)
	}
}

// TestNewUser ensures that authentication works properly.
func TestNewUser(t *testing.T) {
	WithUser(t, func(user *User) {
		if err := user.Authenticate(testPassword); err != nil {
			t.Error("User.Authenticate: got error:\n", err)
		} else if err := user.Authenticate("wrong"); err == nil {
			t.Error("User.Authenticate(\"wrong\"): expected error")
		}

// TestUserValidation ensures that fields are validated by NewUser, SetPassword
// and Save.
func TestValidation(t *testing.T) {
	expectInvalid := func(name, which string, err error) {
		if err == nil {
			t.Errorf("%s: expected error with invalid %s", name, which)
		} else if err, ok := err.(*ErrInvalid); !ok {
			t.Errorf("%s: expected error of type ErrInvalid with invalid user %s", name, which)
		} else if err.Model != "user" {
			t.Errorf("%s: expected error Model field to be 'user' got '%s'", name, err.Model)
		} else if err.Which != which {
			t.Errorf("%s: expected error Which field to be '%s' got '%s'", name, which, err.Which)
		}
	}

	_, err := NewUser("", "test@test.test", testPassword)
	expectInvalid("NewUser", "name", err)

	WithUser(t, func(user *User) {
		user.Email = "test"
		err := user.Save()
		expectInvalid("User.Save", "email", err)

		err = user.SetPassword("bad")
		expectInvalid("User.SetPassword", "password", err)
	})
}

// TestListUser ensures that ListUser returns expected results.
func TestListUser(t *testing.T) {
	if _, err := ListUser(); err == nil {
		t.Error("ListUser: expected error with empty user table")
	} else if _, ok := err.(*ErrEmpty); !ok {
		t.Error("Listuser: expected error of type ErrEmpty, got:\n", err)
	}

	WithUser(t, func(user *User) {
		if err := user.Save(); err != nil {
			t.Error("User.Save: got error:\n", err)
		} else {
			if users, err := ListUser(); err != nil {
				t.Error("ListUser: got error:\n", err)
			} else if len(users) != 1 {
				t.Errorf("ListUser: expected 1 user got %d", len(users))
			} else if users[0].Name != user.Name {
				t.Errorf("ListUser[0].Name: got '%s' expected '%s'", users[0].Name, user.Name)
			}

			if err := user.Delete(); err != nil {
				t.Fatal("User.Delete: got error:\n", err)
			}
		}
	})
}

// TestUserCRUD ensures that basic CRUD operations perform as expected.
func TestUserCRUD(t *testing.T) {
	WithUser(t, func(user *User) {
		now := core.Time()
		created := user.Created
		modified := user.Modified

		if !created.Before(now) {
			t.Errorf("User.Created: expected time to before %s, got %s", now, created)
		}

		if !modified.After(created) && !modified.Equal(created) {
			t.Errorf("User.Modified: expected time after %s, got %s", created, modified)
		}

		if err := user.Delete(); err == nil {
			t.Error("User.Delete: expected error with unsaved user")
		} else if _, ok := err.(*ErrBadEffect); !ok {
			t.Error("User.Delete: expected error of type ErrBadEffect")
		}

		if err := user.Save(); err != nil {
			t.Error("User.Save: got error with new user:\n", err)
		}

		user.Name = "Johnathan Doe"
		if err := user.Save(); err != nil {
			t.Error("User.Save: got error with existing user:\n", err)
		}

		if !user.Modified.After(modified) {
			t.Errorf("User.Modified: expected time to be after %s, got %s", modified, user.Modified)
		}

		if got, err := GetUser(user.Email); err != nil {
			t.Errorf("GetUser(\"%s\"): got error:\n%s", user.Email, err)
		} else if got.Name != user.Name {
			t.Errorf("GetUser.Name: got '%s' expected '%s'", got.Name, user.Name)
		} else if !user.Modified.Equal(got.Modified) {
			t.Errorf("GetUser.Modified: got %s expected %s", got.Modified, user.Modified)
		}

		if err := user.Delete(); err != nil {
			t.Error("User.Delete: got error:\n", err)
		}

		if _, err := GetUser(user.Email); err == nil {
			t.Error("GetUser: expected error with deleted user")
		} else if _, ok := err.(*ErrNoEntry); !ok {
			t.Error("GetUser: expected error of type ErrNoEntry")
		}
	})
}
