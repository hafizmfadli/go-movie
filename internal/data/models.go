package data

import (
	"database/sql"
	"errors"
	"time"
)

var (
	// ErrRecordNotFound is custom error. We'll return this from our Get() method
	// when looking up a movie that doesn't exist in our database
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

// Models is 'container' which can hold and respresent all your database models
type Models struct {
	Movies interface {
		Insert(movie *Movie) error
		Get(id int64) (*Movie, error)
		Update(movie *Movie) error
		Delete(id int64) error
		GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error)
	}
	Users interface {
		Insert(user *User) error
		GetByEmail(email string) (*User, error)
		Update(user *User) error
		GetForToken(tokenScope, tokenPlaintext string) (*User, error)
	}
	Tokens interface {
		Insert(token *Token) error
		DeleteAllForUser(scope string, userID int64) error
		New(userID int64, ttl time.Duration, scope string) (*Token, error)
	}
	Permissions interface {
		GetAllForUser(userID int64) (Permissions, error)
		AddForUser(userID int64, code ...string) error
	}
}

// NewModels return a Models struct
func NewModels(db *sql.DB) Models {
	return Models{
		Movies:      MovieModel{DB: db},
		Users:       UserModel{DB: db},
		Tokens:      TokenModel{DB: db},
		Permissions: PermissionModel{DB: db},
	}
}
