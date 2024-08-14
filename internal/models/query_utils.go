package models

import (
	"math"
	"strings"

	"go.api.template/internal/validator"
)

type Filters struct {
	Page          int
	PageSize      int
	SortColumn    string
	SortDirection string
	SortSafelist  []string
}

// init filters instance
func InitFilters(page int, pageSize int, sortField string, sortSafelist []string) *Filters {
	return &Filters{
		Page:          page,
		PageSize:      pageSize,
		SortColumn:    sortColumn(sortField),
		SortDirection: sortDirection(sortField),
		SortSafelist:  sortSafelist,
	}
}

// return the sort column
func sortColumn(sortField string) string {
	return strings.TrimPrefix(sortField, "-")
}

// return the sort direction ("ASC" or "DESC") depending on the prefix character "-"
func sortDirection(sortField string) string {
	if strings.HasPrefix(sortField, "-") {
		return "DESC"
	}

	return "ASC"
}

// return LIMIT value for pagination
func (f *Filters) limit() int {
	return f.PageSize
}

// return OFFSET value for pagination
func (f *Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

// validate input filters fields
func (f *Filters) ValidateFilters(v *validator.Validator) {
	v.Check(validator.MinNumber(f.Page, 1), "page", "must be greater than zero")
	v.Check(validator.MaxNumber(f.Page, 10_000_000), "page", "must be a maximum of 10 million")
	v.Check(validator.MinNumber(f.PageSize, 1), "page_size", "must be greater than zero")
	v.Check(validator.MaxNumber(f.PageSize, 100), "page_size", "must be a maximum of 100")

	v.Check(validator.PermittedValue(f.SortColumn, f.SortSafelist...), "sort", "invalid sort value")
}

type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

// init metadata instance
func InitMetadata(totalRecords, page, pageSize int) *Metadata {
	if totalRecords == 0 {
		return &Metadata{}
	}

	return &Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}
