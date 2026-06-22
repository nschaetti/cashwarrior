package db

import "fmt"

// SQLFilter is a filter that can be converted to SQL.
type SQLFilter interface {
	GenerateSQL() (condition string, args []any)
}

// Filter is a generic filter that can be used to filter a list of items.
type Filter[T any] interface {
	Reject(T) bool
}

// IDFilter is used to filter a list of items by their ID.
type IDFilter struct {
	ID int64
}

func (f IDFilter) GenerateSQL() (string, []any) {
	return "id = ?", []any{f.ID}
}
func (f IDFilter) String() string {
	return fmt.Sprintf("<IDFilter: %d>", f.ID)
}

// StringFilter is used to filter a list of items by a string.
type StringFilter struct {
	Name  string
	Value string
}

func (f StringFilter) GenerateSQL() (string, []any) {
	return "name LIKE ?", []any{"%" + f.Value + "%"}
}
func (f StringFilter) String() string {
	return fmt.Sprintf("<StringFilter %s:%s>", f.Name, f.Value)
}

func ApplyFilter[T any](items []T, runFilter Filter[T]) []T {
	result := make([]T, 0)
	for _, item := range items {
		if !runFilter.Reject(item) {
			result = append(result, item)
		}
	}
	return result
}
