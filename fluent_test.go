package sqlj

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestFrom(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")

	defer db.Close()

	db.Exec(`
    CREATE TABLE user (id integer primary key, name text, email text, created_at timestamp);
  `)

	if err != nil {
		t.Fatalf("Failed to open db: %s\n", err.Error())
	}

	jdb := DB{
		DB:           db,
		SkipOnInsert: []string{"id"},
	}

	userA := User{
		Name:  "Jess",
		Email: "jess@example.com",
	}

	userB := User{
		Name:  "Joe",
		Email: "joe@example.com",
	}

	userC := User{
		Name:  "Jane",
		Email: "jane@example.com",
	}

	if err := jdb.Insert("user", &userA); err != nil {
		t.Fatalf("Failed to insert user: %s\n", err.Error())
	}

	if err := jdb.Insert("user", &userB); err != nil {
		t.Fatalf("Failed to insert user: %s\n", err.Error())
	}

	if err := jdb.Insert("user", &userC); err != nil {
		t.Fatalf("Failed to insert user: %s\n", err.Error())
	}

	var foundUserA User
	if err := jdb.From("user").Get(userA.ID, &foundUserA); err != nil {
		t.Fatalf("Fluent API, failed to get user: %s\n", err.Error())
	}

	if foundUserA.Name != userA.Name {
		t.Fatalf("Fluent API, name mismatch. Expected: %s, Got: %s\n", userA.Name, foundUserA.Name)
	}

	var foundUserB User
	if err := jdb.From("user").Where("name = ?", userB.Name).One(&foundUserB); err != nil {
		t.Fatalf("Fluent API, failed to get user: %s\n", err.Error())
	}

	if foundUserB.ID != userB.ID {
		t.Fatalf("Fluent API, ID mismatch. Expected: %d, Got: %d\n", userA.ID, foundUserA.ID)
	}

	var allUsers []User
	if err := jdb.From("user").All(&allUsers); err != nil {
		t.Fatalf("Fluent API, failed to select all: %s\n", err.Error())
	}

	if len(allUsers) != 3 {
		t.Fatalf("Fluent API, did not retrieve all users. Expected 3, Got: %d\n", len(allUsers))
	}

	var sortedUsers []User
	if err := jdb.From("user").Order("name", "ASC").All(&sortedUsers); err != nil {
		t.Fatalf("Fluent API, failed to select all: %s\n", err.Error())
	}

	if sortedUsers[0].Name != "Jane" {
		t.Fatalf("Fluent API, sorting failed")
	}

	var reverseSortedUsers []User
	if err := jdb.From("user").Order("name", "DESC").All(&reverseSortedUsers); err != nil {
		t.Fatalf("Fluent API, failed to select all: %s\n", err.Error())
	}

	if reverseSortedUsers[0].Name != "Joe" {
		t.Fatalf("Fluent API, sorting failed")
	}

	count, err := jdb.From("user").Count()

	if err != nil {
		t.Fatalf("Fluent API, failed to count users: %s\n", err.Error())
	}

	if count != 3 {
		t.Fatalf("Fluent API, incorrect count. Expected: 3, Got: %d\n", count)
	}
}
