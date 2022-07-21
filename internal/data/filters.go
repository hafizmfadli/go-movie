package data

import (
	"math"
	"strings"

	"github.com/hafizmfadli/go-movie/internal/validator"
)

type Filters struct {
	Page int
	PageSize int
	Sort string
	SortSafelist []string
}

// sortColumn check the client-provided Sort field matches one of the entries
// in our safelist and if it does, extract the column name from the Sort field
// by stripping the leading hypen character (if one exists).
func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + f.Sort)
}

// sortDirection return the sort direction ("ASC" or "DESC") depending on the prefix character
// of the Sort field.
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC" 
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

// ValidateFilters validate filters value to conform business rules.
// For each invalid filters value will be added as an error to v with
// corresponding key and appropriate message.
func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000,  "page", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")
	v.Check(validator.In(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}

// Metadata struct for holding the pagination metadata.
type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

// calculateMetadata calculates the appropriate pagination metadata
// values given the total number of records, current page, and page size values.
// Note that the last page value is calculated using the math.Ceil() function,
// which rounds up a float to the nearest integer. So, for example, if there were 12
// records in toal and a page size of 5, the last page value would be math.Ceil(12/5) = 3
func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage: page,
		PageSize: pageSize,
		FirstPage: 1,
		LastPage: int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}