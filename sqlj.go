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

// Gets a single row from the given table with the given id.
// v must be a pointer to a struct.
func (jdb *DB) Get(table string, id any, v any) error {
	if err := checkValueType(v); err != nil {
		return err
	}

	fields := extractFields(v)
	columns := pluckNames(fields)

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
	fields := extractFields(structInstance)
	columns := pluckNames(fields)

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

type Options struct {
	Fields []Field
}

// Inserts a row into the specified `table` with the given struct.
// The new row is returned and marshalled into v.
// v must be a pointer to a struct.
func (jdb *DB) Insert(table string, v any) error {
	return jdb.InsertWithOptions(table, Options{}, v)
}

// Inserts a row into the specified `table` with the given struct.
// The new row is returned and marshalled into v.
// An Options type with a slice of Fields can be included to override any values in v.
// v must be a pointer to a struct.
func (jdb *DB) InsertWithOptions(table string, options Options, v any) error {
	if err := checkValueType(v); err != nil {
		return err
	}

	allFields := extractFields(v)
	fields := append(allFields, options.Fields...)
	fields = dedupeFields(fields)

	filteredFields := filterFields(fields, jdb.SkipOnInsert)
	returnColumns := pluckNames(allFields)

	sql := buildInsertSQL(table, filteredFields, returnColumns)

	values := pluckValues(filteredFields)

	return jdb.GetRow(sql, v, values...)
}

// Updates a row in the specified `table` using the given struct.
// The updated row is returned and marshalled into v.
// v must be a pointer to a struct.
func (jdb *DB) Update(table string, id any, v any) error {
	if err := checkValueType(v); err != nil {
		return err
	}

	fields := extractFields(v)
	filteredFields := filterFields(fields, jdb.SkipOnInsert)
	returnColumns := pluckNames(fields)

	sql := buildUpdateSQL(table, filteredFields, returnColumns)

	values := pluckValues(filteredFields)
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

func extractFields(v any) []Field {
	t := reflect.TypeOf(v)

	number_of_fields := t.Elem().NumField()
	value := reflect.ValueOf(v).Elem()

	fields := make([]Field, number_of_fields)

	n := 0
	for i := 0; i < number_of_fields; i++ {
		dbTag := t.Elem().Field(i).Tag.Get("db")

		if dbTag == "" || dbTag == "-" {
			continue
		}

		fields[n] = BasicField{
			Name:  dbTag,
			Value: value.Field(i).Addr().Interface(),
		}
		n++
	}

	return fields[:n]
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

func buildInsertSQL(table string, fields []Field, columns []string) string {
	names := make([]string, len(fields))
	placeholders := make([]string, len(fields))

	for idx, f := range fields {
		names[idx] = f.GetName()
		placeholders[idx] = f.GetPlaceholder(idx)
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

func buildUpdateSQL(table string, fields []Field, columns []string) string {
	setExpressions := make([]string, len(fields))
	for idx, f := range fields {
		setExpressions[idx] = fmt.Sprintf("%s = %s", f.GetName(), f.GetPlaceholder(idx))
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
