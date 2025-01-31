package sqlj

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

type DBLike interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

type DB struct {
	DB           DBLike
	SkipOnInsert []string // Allows you specify db field names to skip on insert
}

type field struct {
	Name        string
	Placeholder string
	Value       any
}

// Gets a single row from the given table with the given id.
// v must be a pointer to a struct.
func (jdb *DB) Get(table string, id any, v any) error {
	if err := checkValueType(v); err != nil {
		return err
	}

	columns := extractColumns(v)

	sql := strings.Join([]string{"SELECT ", strings.Join(columns, ", "), " FROM ", table, " WHERE id = $1"}, "")

	return jdb.GetRow(sql, v, id)
}

// Gets a single row using the supplied SQL and values.
// The result will be marshalled into the v struct.
// v must be a pointer to a struct.
func (jdb *DB) GetRow(sql string, v any, values ...any) error {
	row := jdb.DB.QueryRow(sql, values...)

	return scanIntoStruct(row, v)
}

// Selects all rows from a given table.
// The results will be marshalled into the v slice of structs.
// v must be a pointer to a slice of structs.
func (jdb *DB) Select(table string, v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Slice {
		return errors.New("v must be a pointer to a slice of structs")
	}

	structType := val.Elem().Type().Elem()
	if structType.Kind() != reflect.Struct {
		return errors.New("v must be a pointer to a slice of structs")
	}

	structInstance := reflect.New(structType).Interface()
	columns := extractColumns(structInstance)

	sql := strings.Join([]string{"SELECT ", strings.Join(columns, ", "), " FROM ", table}, "")

	return jdb.SelectAll(sql, v)
}

// Selects all rows using the supplied SQL and values.
// The results will be marshalled into the v slice of structs.
// v must be a pointer to a slice of structs.
func (jdb *DB) SelectAll(sql string, v any, values ...any) error {
	rows, err := jdb.DB.Query(sql, values...)

	if err != nil {
		return err
	}

	return scanRowsIntoStructs(rows, v)
}

// Inserts a row into the specified `table` with the given struct.
// The new row is returned and marshalled into v.
// v must be a pointer to a struct.
func (jdb *DB) Insert(table string, v any) error {
	if err := checkValueType(v); err != nil {
		return err
	}

	fields := extractFields(v, jdb.SkipOnInsert)
	columns := extractColumns(v)
	sql := buildInsertSQL(table, fields, columns)

	values := make([]any, len(fields))
	for idx, f := range fields {
		values[idx] = f.Value
	}

	return jdb.GetRow(sql, v, values...)
}

// Updates a row in the specified `table` using the given struct.
// The updated row is returned and marshalled into v.
// v must be a pointer to a struct.
func (jdb *DB) Update(table string, id any, v any) error {
	if err := checkValueType(v); err != nil {
		return err
	}

	fields := extractFields(v, jdb.SkipOnInsert)
	columns := extractColumns(v)
	sql := buildUpdateSQL(table, fields, columns)

	values := make([]any, len(fields))
	for idx, f := range fields {
		values[idx] = f.Value
	}

	values = append(values, id)

	return jdb.GetRow(sql, v, values...)
}

// Deletes a row in the given table by ID.
func (jdb *DB) Delete(table string, id any) error {
	sql := strings.Join([]string{"DELETE FROM ", table, " WHERE id = $1"}, "")

	// TODO: It would be prudent to check RowsAffected() on the result.
	// I need to look into how this is supports with different DB drivers.
	_, err := jdb.DB.Exec(sql, id)

	return err
}

func scanIntoStruct(row *sql.Row, dest any) error {
	val := reflect.ValueOf(dest)

	columns := make([]interface{}, val.Elem().NumField())
	for i := 0; i < val.Elem().NumField(); i++ {
		columns[i] = val.Elem().Field(i).Addr().Interface()
	}

	return row.Scan(columns...)
}

func scanRowsIntoStructs(rows *sql.Rows, dest interface{}) error {
	val := reflect.ValueOf(dest)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Slice {
		return errors.New("dest must be a pointer to a slice of structs")
	}

	structType := val.Elem().Type().Elem()
	if structType.Kind() != reflect.Struct {
		return errors.New("dest must be a pointer to a slice of structs")
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	for rows.Next() {
		structInstance := reflect.New(structType).Interface()

		fieldPointers := make([]interface{}, len(columns))
		for i := 0; i < len(columns); i++ {
			fieldPointers[i] = reflect.ValueOf(structInstance).Elem().Field(i).Addr().Interface()
		}

		if err := rows.Scan(fieldPointers...); err != nil {
			return err
		}

		val.Elem().Set(reflect.Append(val.Elem(), reflect.ValueOf(structInstance).Elem()))
	}

	return rows.Err()
}

func extractColumns(v any) []string {
	fields := extractFields(v, []string{})

	columns := make([]string, len(fields))

	for idx, f := range fields {
		columns[idx] = f.Name
	}

	return columns
}

func extractFields(v any, skipColumns []string) []field {
	t := reflect.TypeOf(v)

	number_of_fields := t.Elem().NumField()
	value := reflect.ValueOf(v).Elem()

	fields := make([]field, number_of_fields)

	n := 0
	for i := 0; i < number_of_fields; i++ {
		dbTag := t.Elem().Field(i).Tag.Get("db")

		if dbTag == "" || dbTag == "-" || slices.Contains(skipColumns, dbTag) {
			continue
		}

		fields[n] = field{
			Name:        dbTag,
			Placeholder: fmt.Sprintf("$%d", n),
			Value:       value.Field(i).Addr().Interface(),
		}
		n++
	}

	return fields[:n]
}

func buildInsertSQL(table string, fields []field, columns []string) string {
	names := make([]string, len(fields))
	placeholders := make([]string, len(fields))

	for idx, f := range fields {
		names[idx] = f.Name
		placeholders[idx] = f.Placeholder
	}

	return strings.Join(
		[]string{
			"INSERT INTO ",
			table,
			" (",
			strings.Join(names, ", "),
			") VALUES (",
			strings.Join(placeholders, ", "),
			") RETURNING ",
			strings.Join(columns, ", "),
		},
		"",
	)
}

func buildUpdateSQL(table string, fields []field, columns []string) string {
	setExpressions := make([]string, len(fields))
	for idx, f := range fields {
		setExpressions[idx] = fmt.Sprintf("%s = %s", f.Name, f.Placeholder)
	}

	return strings.Join(
		[]string{
			"UPDATE ",
			table,
			" SET ",
			strings.Join(setExpressions, ", "),
			fmt.Sprintf(" WHERE id = $%d ", len(fields)),
			"RETURNING ",
			strings.Join(columns, ", "),
		},
		"",
	)
}

func checkValueType(v any) error {
	t := reflect.TypeOf(v)

	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		return errors.New("Value must be a pointer to a struct")
	}

	return nil
}
