package data

import (
	"database/sql"
	"errors"
)

var (
	// ErrRecordNotFound is custom error. We'll return this from our Get() method
	// when looking up a movie that doesn't exist in our database
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict = errors.New("edit conflict")
)

// Models is 'container' which can hold and respresent all your database models
type Models struct {
	Movies interface {
		Insert(movie *Movie) error
		Get(id int64) (*Movie, error)
		Update(movie *Movie) (error)
		Delete(id int64) error
	}
}

// NewModels return a Models struct
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}