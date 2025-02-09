package sqlj

import (
	"fmt"
	"slices"
)

type Field interface {
	GetName() string
	GetValue() any

	// For an insert this is the $1 in the VALUES list.
	// For an update this is the $1 in the SET expression.
	GetPlaceholder(idx int) string

	// This isn't ideal but will work for now
	IsLiteral() bool
}

// This is a standard k = v field
type BasicField struct {
	Name  string
	Value any
}

func (f BasicField) GetName() string {
	return f.Name
}

func (f BasicField) GetValue() any {
	return f.Value
}

func (f BasicField) GetPlaceholder(idx int) string {
	return fmt.Sprintf("$%d", idx)
}

func (f BasicField) IsLiteral() bool {
	return false
}

// This is useful if you want to call a function.
// An example would be LiteralField{Name: "created_at", Value: "now()"}.
type LiteralField struct {
	Name  string
	Value string
}

func (f LiteralField) GetName() string {
	return f.Name
}

func (f LiteralField) GetValue() any {
	return nil
}

func (f LiteralField) GetPlaceholder(idx int) string {
	return f.Value
}

func (f LiteralField) IsLiteral() bool {
	return true
}

func pluckNames(fields []Field) []string {
	names := make([]string, len(fields))

	for idx, f := range fields {
		names[idx] = f.GetName()
	}

	return names
}

func pluckValues(fields []Field) []any {
	values := make([]any, len(fields))

	n := 0
	for _, f := range fields {
		if v := f.GetValue(); v != nil {
			values[n] = v
			n++
		}
	}

	return values[:n]
}

// TODO: Rewrite this so it's deterministic. Currently the order the fields is changed.
func dedupeFields(fields []Field) []Field {
	indexedFields := make(map[string]Field)

	for _, f := range fields {
		indexedFields[f.GetName()] = f
	}

	allFields := make([]Field, len(indexedFields))

	n := 0
	for _, v := range indexedFields {
		allFields[n] = v
		n++
	}

	return allFields
}

func filterFields(fields []Field, skipColumns []string) []Field {
	outFields := make([]Field, len(fields))

	n := 0
	for _, f := range fields {
		if slices.Contains(skipColumns, f.GetName()) {
			continue
		}

		outFields[n] = f
		n++
	}

	return outFields[:n]
}
