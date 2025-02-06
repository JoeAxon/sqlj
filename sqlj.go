package sqlj

import (
	"database/sql"
	"errors"
)

type DB struct {
	DB           DBLike
	IDColumn     string
	SkipOnInsert []string // Allows you specify db field names to skip on insert
}

// Represents a DB-like interface. This only specifies the methods used by sqlj.
// Both DB and Tx in the database/sql standard library fulfill this contract.
type DBLike interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// Gets a single row from the given table with the given id.
// v must be a pointer to a struct.
func (jdb *DB) Get(table string, id any, v any) error {
	if err := checkValueType(v); err != nil {
		return err
	}

	fields := extractFields(v)
	columns := pluckNames(fields)

	sql := buildSelectQuery(Select{
		Columns: columns,
		From:    table,
		Where: []WhereClause{
			{AND_TYPE, SimpleExpr{columnEq(jdb.GetIDName())}},
		},
	})

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
	structInstance, err := getSliceStructInstance(v)
	if err != nil {
		return err
	}

	fields := extractFields(structInstance)
	columns := pluckNames(fields)

	sql := buildSelectQuery(Select{
		Columns: columns,
		From:    table,
	})

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

type PageOptions struct {
	pageNumber uint
	pageSize   uint
	order      []OrderBy
}

type OrderBy struct {
	expression string
	direction  string
}

// Selects a page of data from the given table.
// The options parameter allows you to specify the page, page size and order by clauses.
// The results will be marshalled into the v slice of structs.
// v must be a pointer to a slice of structs.
func (jdb *DB) Page(table string, options PageOptions, v any) error {
	if options.pageNumber < 1 {
		return errors.New("Page number must be greater than 0")
	}

	if options.pageSize < 1 {
		return errors.New("Page size must be greater than 0")
	}

	if len(options.order) == 0 {
		return errors.New("Must include atleast one order by")
	}

	structInstance, err := getSliceStructInstance(v)
	if err != nil {
		return err
	}

	fields := extractFields(structInstance)
	columns := pluckNames(fields)

	offset := (options.pageNumber - 1) * options.pageSize
	limit := options.pageSize

	sql := buildSelectQuery(Select{
		From:    table,
		OrderBy: options.order,
		Columns: columns,
		Offset:  true,
		Limit:   true,
	})

	return jdb.SelectAll(sql, v, offset, limit)
}

// Counts the number of records in the table.
// This is intended to be used in conjunction with .Page.
func (jdb *DB) Count(table string) (uint, error) {
	var count uint = 0

	sql := buildSelectQuery(Select{
		Columns: []string{"count(1)"},
		From:    table,
	})

	result := jdb.DB.QueryRow(sql)

	if err := result.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
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

	sql := buildInsertSQL(Insert{
		From:      table,
		Fields:    filteredFields,
		Returning: returnColumns,
	})

	values := pluckValues(filteredFields)

	return jdb.GetRow(sql, v, values...)
}

// Updates a row in the specified `table` using the given struct.
// The updated row is returned and marshalled into v.
// v must be a pointer to a struct.
func (jdb *DB) Update(table string, id any, v any) error {
	return jdb.UpdateWithOptions(table, id, Options{}, v)
}

// Updates a row in the specified `table` using the given struct.
// The updated row is returned and marshalled into v.
// An Options type with a slice of Fields can be included to override any values in v.
// v must be a pointer to a struct.
func (jdb *DB) UpdateWithOptions(table string, id any, options Options, v any) error {
	if err := checkValueType(v); err != nil {
		return err
	}

	allFields := extractFields(v)
	fields := append(allFields, options.Fields...)
	fields = dedupeFields(fields)

	filteredFields := filterFields(fields, jdb.SkipOnInsert)
	returnColumns := pluckNames(allFields)

	sql := buildUpdateSQL(table, filteredFields, returnColumns)

	values := pluckValues(filteredFields)
	values = append(values, id)

	return jdb.GetRow(sql, v, values...)
}

// Deletes a row in the given table by ID.
func (jdb *DB) Delete(table string, id any) error {
	sql := buildDeleteSQL(Delete{
		From: table,
		Where: []WhereClause{
			{AND_TYPE, SimpleExpr{columnEq(jdb.GetIDName())}},
		},
	})

	// TODO: It would be prudent to check RowsAffected() on the result.
	// I need to look into how this is supports with different DB drivers.
	_, err := jdb.DB.Exec(sql, id)

	return err
}

func (jdb *DB) GetIDName() string {
	if jdb.IDColumn == "" {
		return "id"
	}

	return jdb.IDColumn
}
