package sqlj

import "fmt"

type Field interface {
	GetName() string
	GetValue() any
	GetPlaceholder(idx int) string
}

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

func pluckNames(fields []Field) []string {
	names := make([]string, len(fields))

	for idx, f := range fields {
		names[idx] = f.GetName()
	}

	return names
}

func pluckValues(fields []Field) []any {
	values := make([]any, len(fields))

	for idx, f := range fields {
		if v := f.GetValue(); v != nil {
			values[idx] = v
		}
	}

	return values
}
