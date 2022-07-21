package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/hafizmfadli/go-movie/internal/validator"
	"github.com/lib/pq"
)

type Movie struct {
	// Unique ID
	ID        int64     `json:"id"`
	// Timestamp for when the movie is added to our database
	CreatedAt time.Time `json:"-"`
	// Movie title
	Title     string    `json:"title"`
	// Movie release year
	Year      int32     `json:"year,omitempty"`   
	// Movie runtime (in minutes)    
	Runtime   int32     `json:"runtime,omitempty"`
	// Slice of genres for the movie (romance, comedy, etc.)
	Genres    []string  `json:"genres,omitempty"`
	// Version number starts at 1 and will be incremented each time the movie is updated
	Version   int32     `json:"version"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")
	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")
	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain dulicate values")
}

type MovieModel struct {
	DB *sql.DB
}

// Insert inserting a new record in the movies table
func (m MovieModel) Insert(movie *Movie) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	query := `INSERT INTO movies (title, year, runtime, genres) VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version`

	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}
	
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// Get fetching a specific record from the movies table
func (m MovieModel) Get(id int64) (*Movie, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT id, created_at, title, year, runtime, genres, version
	FROM movies WHERE id = $1`

	movie := Movie{}

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(&movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, 
	&movie.Runtime, pq.Array(&movie.Genres), &movie.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

// Update updating a specific record in the movies table
func (m MovieModel) Update(movie *Movie) (error) {
	query := `UPDATE movies 
	SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
	WHERE id = $5 AND version = $6
	RETURNING version`

	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres), movie.ID, movie.Version}

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

// Delete deleting a specific record from the movies table
func (m MovieModel) Delete(id int64) error {

	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM movies WHERE id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	
	return nil
}

func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error) {
	
	// notice that we also include a secondary sort on the movie ID to ensure a
	// consistent ordering.
	query := fmt.Sprintf(`
	SELECT count(*) OVER(), id, created_at, title, year, runtime, genres, version
	FROM movies
	WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
	AND (genres @> $2 OR $2 = '{}')
	ORDER BY %s %s, id ASC
	LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	args := []interface{}{title, pq.Array(genres), filters.limit(), filters.offset()}

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	movies := []*Movie{}
	totalRecords := 0

	for rows.Next() {
		var m Movie

		err = rows.Scan(
			&totalRecords,
			&m.ID, 
			&m.CreatedAt, 
			&m.Title, 
			&m.Year, 
			&m.Runtime, 
			pq.Array(&m.Genres), 
			&m.Version)
		
		if err != nil {
			return nil, Metadata{}, err
		}
		
		movies = append(movies, &m)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return movies, metadata, nil
}