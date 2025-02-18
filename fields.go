package sqlj

import (
	"fmt"
	"slices"
)

type field interface {
	GetName() string
	GetValue() any

	// For an insert this is the $1 in the VALUES list.
	// For an update this is the $1 in the SET expression.
	GetPlaceholder(idx int) string

	// This isn't ideal but will work for now
	IsLiteral() bool
}

// This is a standard k = v field
type basicField struct {
	Name  string
	Value any
}

func (f basicField) GetName() string {
	return f.Name
}

func (f basicField) GetValue() any {
	return f.Value
}

func (f basicField) GetPlaceholder(idx int) string {
	return fmt.Sprintf("$%d", idx)
}

func (f basicField) IsLiteral() bool {
	return false
}

// This is useful if you want to call a function.
// An example would be literalField{Name: "created_at", Value: "now()"}.
type literalField struct {
	Name  string
	Value string
}

func (f literalField) GetName() string {
	return f.Name
}

func (f literalField) GetValue() any {
	return nil
}

func (f literalField) GetPlaceholder(idx int) string {
	return f.Value
}

func (f literalField) IsLiteral() bool {
	return true
}

func pluckNames(fields []field) []string {
	names := make([]string, len(fields))

	for idx, f := range fields {
		names[idx] = f.GetName()
	}

	return names
}

func pluckValues(fields []field) []any {
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
func dedupeFields(fields []field) []field {
	indexedFields := make(map[string]field)

	for _, f := range fields {
		indexedFields[f.GetName()] = f
	}

	allFields := make([]field, len(indexedFields))

	n := 0
	for _, v := range indexedFields {
		allFields[n] = v
		n++
	}

	return allFields
}

func filterFields(fields []field, skipColumns []string) []field {
	outFields := make([]field, len(fields))

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
