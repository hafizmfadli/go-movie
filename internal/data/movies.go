package data

import "time"

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
