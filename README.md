# sqlj

[![Go Reference](https://pkg.go.dev/badge/github.com/JoeAxon/sqlj.svg)](https://pkg.go.dev/github.com/JoeAxon/sqlj)

A simple struct to record mapper. Builds on top of the database/sql standard library to provide basic CRUD operations for flat structs.

*Note:* sqlj is in active development and should not be relied upon in production.

## Install

```
go get github.com/JoeAxon/sqlj
```

## Basic Usage

The library is intentionally limited with a few escape hatches for more complicated use cases. It is not intended to be an ORM but merely a convenient way to insert, update and select records from a database with minimal ceremony. To this end there is no support for marshalling records into nested structs.

### Setup

The simplest way to get started with sqlj is to provide the `Open` function with the driver name and data source name:

```go
package main

import (
	"fmt"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/JoeAxon/sqlj"
)

func main() {
	// .Open accepts the same arguments as DB.Open in the database/sql package
	db, err := sqlj.Open("sqlite3", ":memory:")

	// If you are using PostgreSQL with lib/pg, this might look like:
	// db, err = sql.Open("postgres", "host=localhost user=youruser password=yourpassword dbname=yourdb port=5432")

	defer db.Close()
}
```

The `NewDB` function can also be used to set-up sqlj. This accepts any struct that implements the `DBLike` interface. This is useful if you need to open the connection to the DB separately using the standard library database/sql package or if you are working with transactions.

You can also initialise the `sqlj.DB` manually if you find it useful to do so. I don't think it is necessary to do but I also don't intend to make it problematic if you do.

### Inserting and updating records

Inserting and updating is straightforward with sqlj. As long as your structs are tagged with the corresponding `db` field name in the database you should just be able to call the `Insert` and `Update` methods. It's worth noting that the structs don't have to map every field in table. For `Insert`, only the `NOT NULL` fields need to be present to generate valid SQL and for `Update` you're free to omit whichever fields you like. This can be a powerful pattern where you create types for specific mutations.

```go

// The "db" struct tags specify the column name in the database.
// Only fields with a "db" tag will be included in queries.
type User struct {
	ID    uint   `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

func main() {
	db, err := sqlj.Open("sqlite3", ":memory:")
	defer db.Close()

	user := User{Name: "Joe", Email: "joe@example.com"}

	// .Insert will generate and execute an "INSERT" query on the "users" table using the struct provided.
	// The newly created user will be unmarshalled into the struct.
	if err := db.Insert("users", &user); err != nil {
		fmt.Fatalf("Failed to insert user: %s\n", err.Error())
	}

	user.Name = "John"

	// .Update will generate an "UPDATE" statement on the "users" table using the struct provided.
	if err := db.Update("users", user.ID, &user); err != nil {
		fmt.Fatalf("Failed to update user: %s\n", err.Error())
	}
}
```

You can get a little more control over the generated SQL by using the `InsertWithFields` and `UpdateWithFields` methods. A real world example might be setting the `updated_at` field on a record to the current timestamp:

```go
// The map keys should match a column in the table and the key a literal value.
// The keys and values are simply interpolated into the "UPDATE" query so be careful
// when passing dynamic strings to this method not to introduce an opportunity for SQL injection.
if err := db.UpdateWithFields("users", &user, map[string]string{"updated_at": "date()"}); err != nil {
	t.Fatalf("Failed to update user: %s\n", err.Error())
}
```

### Retrieving records

The DB struct exposes the `GetRow` and `SelectAll` functions to allow you to marshall the results of arbitrary SQL into a struct or slice of structs respectively. It also exposes the `Get` function for retrieving a record by ID and, less usefully, the `Select` function to retrieve all records from a table.

```go
var user User

// The .Get function is equivalent to the .GetRow call below
if err := db.Get("users", 1, &user); err != nil {
	fmt.Fatalf("Failed to retrieve user: %s\n", err.Error())
}

if err := db.GetRow("SELECT id, name, email FROM users WHERE id = $1", &user, 1); err != nil {
	fmt.Fatalf("Failed to retrieve user: %s\n", err.Error())
}

var allUsers []User

// The .Select function is equivalent to the .SelectAll call below
if err := db.Select("users", &allUsers); err != nil {
	fmt.Fatalf("Failed to retrieve users: %s\n", err.Error())
}

if err := db.SelectAll("SELECT id, name, email FROM users", &allUsers); err != nil {
	fmt.Fatalf("Failed to retrieve users: %s\n", err.Error())
}
```

### Fluent API

There is an ergonomic API for writing queries that should hopefully suffice in most cases. Fluent interfaces get a bad rap but I believe this is a valid usecase and not too egregious:

```go
var user User

// .From returns a QueryDB struct which allows you to chain .Where,
// .OrWhere and .Order calls before calling the .One method which
// marshalls a single record into the given struct.
if err := db.From("users").Where("name = ?", "Joe").One(&user); err != nil {
	fmt.Fatalf("Failed to retrieve user: %s\n", err.Error())
}

var allJoes []User

// .All will marshall multiple records into a slice of structs.
if err := db.From("users").Where("name = ?", "Joe").All(&allJoes); err != nil {
	fmt.Fatalf("Failed to retrieve all Joes: %s\n", err.Error())
}

var firstPage []User

// .Page will retrieve a page of records given a page number and size.
if err := db.From("users").Order("name", "ASC").Page(1, 10, &firstPage); err != nil {
	fmt.Fatalf("Failed to retrieve first page: %s\n", err.Error())
}

// .Count will return a count of records for the given query.
// This is intended to be used in conjunction with the .Page method.
total, err := db.From("users").Where("name <> ?", "Joe").Count()
if err != nil {
	fmt.Fatalf("Failed to retrieve first page: %s\n", err.Error())
}
```

