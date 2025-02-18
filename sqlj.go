package sqlj

import (
	"database/sql"
	"fmt"
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

func NewDB(db DBLike) DB {
	return DB{
		DB:           db,
		SkipOnInsert: []string{"id"},
	}
}

func Open(driver string, dsn string) (*DB, error) {
	db, err := sql.Open(driver, dsn)

	if err != nil {
		return nil, err
	}

	jdb := NewDB(db)

	return &jdb, nil
}

func (jdb *DB) Close() {
	db, ok := jdb.DB.(*sql.DB)

	if ok {
		db.Close()
	}
}

// Gets a single row from the given table with the given id.
// v must be a pointer to a struct.
func (jdb *DB) Get(table string, id any, v any) error {
	if err := checkValueType(v); err != nil {
		return err
	}

	fields := extractFields(v)
	columns := pluckNames(fields)

	sql := buildSelectQuery(selectParams{
		Columns: columns,
		From:    table,
		Where: []WhereClause{
			{AND_TYPE, SimpleExpr{columnEq(jdb.getIDName())}},
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

	sql := buildSelectQuery(selectParams{
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

// Inserts a row into the specified `table` with the given struct.
// The new row is returned and marshalled into v.
// v must be a pointer to a struct.
func (jdb *DB) Insert(table string, v any) error {
	return jdb.InsertWithFields(table, v, map[string]string{})
}

// Inserts a row into the specified `table` with the given struct.
// The new row is returned and marshalled into v.
// A map of column to literal string value can be included to override any values in v.
// v must be a pointer to a struct.
func (jdb *DB) InsertWithFields(table string, v any, fieldMap map[string]string) error {
	if err := checkValueType(v); err != nil {
		return err
	}

	allFields := extractFields(v)
	literalFields := literalFieldsFromMap(fieldMap)
	fields := append(allFields, literalFields...)
	fields = dedupeFields(fields)

	filteredFields := filterFields(fields, jdb.SkipOnInsert)
	returnColumns := pluckNames(allFields)

	sql := buildInsertSQL(insertParams{
		From:      table,
		Fields:    filteredFields,
		Returning: returnColumns,
	})

	if len(fieldMap) > 0 {
		fmt.Printf("Fields: %v\n", fields)
		fmt.Printf("Insert with fields SQL: %s\n", sql)
	}

	values := pluckValues(filteredFields)

	return jdb.GetRow(sql, v, values...)
}

// Updates a row in the specified `table` using the given struct.
// The updated row is returned and marshalled into v.
// v must be a pointer to a struct.
func (jdb *DB) Update(table string, id any, v any) error {
	return jdb.UpdateWithFields(table, id, v, map[string]string{})
}

// Updates a row in the specified `table` using the given struct.
// The updated row is returned and marshalled into v.
// A map of column to literal string value can be included to override any values in v.
// v must be a pointer to a struct.
func (jdb *DB) UpdateWithFields(table string, id any, v any, fieldMap map[string]string) error {
	if err := checkValueType(v); err != nil {
		return err
	}

	allFields := extractFields(v)
	literalFields := literalFieldsFromMap(fieldMap)
	fields := append(allFields, literalFields...)
	fields = dedupeFields(fields)

	filteredFields := filterFields(fields, jdb.SkipOnInsert)
	returnColumns := pluckNames(allFields)

	sql := buildUpdateSQL(updateParams{
		From:      table,
		Fields:    filteredFields,
		Returning: returnColumns,
	})

	values := pluckValues(filteredFields)
	values = append(values, id)

	return jdb.GetRow(sql, v, values...)
}

// Deletes a row in the given table by ID.
func (jdb *DB) Delete(table string, id any) error {
	sql := buildDeleteSQL(deleteParams{
		From: table,
		Where: []WhereClause{
			{AND_TYPE, SimpleExpr{columnEq(jdb.getIDName())}},
		},
	})

	// TODO: It would be prudent to check RowsAffected() on the result.
	// I need to look into how this is supports with different DB drivers.
	_, err := jdb.DB.Exec(sql, id)

	return err
}

func (jdb *DB) From(table string) QueryDB {
	return QueryDB{
		DB:   jdb,
		From: table,
	}
}

func (jdb *DB) getIDName() string {
	if jdb.IDColumn == "" {
		return "id"
	}

	return jdb.IDColumn
}
