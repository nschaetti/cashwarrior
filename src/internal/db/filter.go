package db

import "fmt"

type Filter interface {
	GenerateSQL() (condition string, args []any)
}

type IDFilter struct {
	ID int64
}

func (f IDFilter) GenerateSQL() (string, []any) {
	return "id = ?", []any{f.ID}
}

func (f IDFilter) String() string {
	return fmt.Sprintf("<IDFilter: %d>", f.ID)
}

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
