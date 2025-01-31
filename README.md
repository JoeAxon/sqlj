# sqlj

[![Go Reference](https://pkg.go.dev/badge/github.com/JoeAxon/sqlj.svg)](https://pkg.go.dev/github.com/JoeAxon/sqlj)

A simple struct to record mapper. Builds on top of the database/sql standard library to provide basic CRUD operations for flat structs.

*Note:* sqlj is in active development and should not be relied upon in production.

## Install

```
go get github.com/JoeAxon/sqlj
```

## Usage

```go
jdb := sqlj.DB{
  DB: db,
}

user := User{}

if err := jdb.Insert("users", &user); err != nil {
  fmt.Printf("Insert failed: %s\n", err.Error())
  return
}
```

## Functions

- `Get` - Gets and unmarshalls a record to a struct by ID.
- `GetRow` - Unmarshalls the result of a SQL query to a struct.
- `Select` - Selects all records in a table / view and unmarshalls to a slice of structs.
- `SelectAll` - Unmarshalls the result of a SQL query to a slice of structs.
- `Insert` - Inserts a struct and unmarshalls the result to the same struct.
- `Update` - Update a record by ID, unmarshalling the result to a struct.
- `Delete` - Deletes a record by ID.
