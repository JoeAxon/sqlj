# sqlj

[![Go Reference](https://pkg.go.dev/badge/github.com/JoeAxon/sqlj.svg)](https://pkg.go.dev/github.com/JoeAxon/sqlj)

A simple struct to record mapper. Builds on top of the database/sql standard library to provide basic CRUD operations for flat structs.

*Note:* sqlj is in active development and should not be relied upon in production.

## Install

```
go get github.com/JoeAxon/sqlj
```

## Basic Usage

```go
package main

import (
    "fmt"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
    "github.com/JoeAxon/sqlj"
)

// The "db" struct tags specify the column name in the database.
// Only fields with a "db" tag will be included in queries.
type User struct {
	ID    uint   `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

func main() {
    // DB setup, this will likely be different for you.
	db, err := sql.Open("sqlite3", ":memory:")
	db.Exec("CREATE TABLE users (id integer primary key, name text, email text)")
	defer db.Close()

    // SkipOnInsert can be omitted
	jdb := DB{
		DB:           db,
		SkipOnInsert: []string{"id"},
	}

	user := User{Name: "Joe", Email: "joe@example.com"}

    // .Insert will generate and execute an "INSERT" query on the "users" table using the struct provided.
    // The newly created user will be unmarshalled into the struct.
	if err := jdb.Insert("users", &user); err != nil {
		fmt.Fatalf("Failed to insert user: %s\n", err.Error())
	}

	foundUser := User{}

	if err := jdb.Get("users", user.ID, &foundUser); err != nil {
		fmt.Fatalf("Failed to retrieve user: %s\n", err.Error())
	}

    fmt.Println("Found the user: ", foundUser)
}
```

Further examples of usage can be found in `sqlj_test.go`.

## Functions

- `Get` - Gets and unmarshalls a record to a struct by ID.
- `GetRow` - Unmarshalls the result of a SQL query to a struct.
- `Select` - Selects all records in a table / view and unmarshalls to a slice of structs.
- `SelectAll` - Unmarshalls the result of a SQL query to a slice of structs.
- `Insert` - Inserts a struct and unmarshalls the result to the same struct.
- `InsertWithOptions` - Same as Insert but also allows fields to be overriden.
- `Update` - Update a record by ID, unmarshalling the result to a struct.
- `UpdateWithOptions` - Same as Update but also allows fields to be overriden.
- `Delete` - Deletes a record by ID.
