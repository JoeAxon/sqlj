package sqlj

import (
	"database/sql"
	"errors"
	"reflect"
)

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

func checkValueType(v any) error {
	t := reflect.TypeOf(v)

	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		return errors.New("Value must be a pointer to a struct")
	}

	return nil
}

func getSliceStructInstance(v any) (any, error) {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Slice {
		return nil, errors.New("v must be a pointer to a slice of structs")
	}

	structType := val.Elem().Type().Elem()
	if structType.Kind() != reflect.Struct {
		return nil, errors.New("v must be a pointer to a slice of structs")
	}

	return reflect.New(structType).Interface(), nil
}
