package sqlj

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID    uint   `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

func TestInsertAndRetrieve(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")

	defer db.Close()

	db.Exec("CREATE TABLE user (id integer primary key, name text, email text)")

	if err != nil {
		t.Fatalf("Failed to open db: %s\n", err.Error())
	}
	userA := User{Name: "Joe", Email: "joe@example.com"}
	userB := User{Name: "Jen", Email: "jen@example.com"}

	jdb := DB{
		DB:           db,
		SkipOnInsert: []string{"id"},
	}

	if err := jdb.Insert("user", &userA); err != nil {
		t.Fatalf("Failed to insert user: %s\n", err.Error())
	}

	if userA.Name != "Joe" || userA.Email != "joe@example.com" {
		t.Fatal("Returned data does not match inserted data\n")
	}

	if err := jdb.Insert("user", &userB); err != nil {
		t.Fatalf("Failed to insert user: %s\n", err.Error())
	}

	foundUserA := User{}

	if err := jdb.Get("user", userA.ID, &foundUserA); err != nil {
		t.Fatalf("Failed to retrieve user: %s\n", err.Error())
	}

	if userA.ID != foundUserA.ID {
		t.Fatal("ID mismatch when retrieving user")
	}

	allUsers := []User{}

	if err := jdb.Select("user", &allUsers); err != nil {
		t.Fatalf("Failed to select all users: %s\n", err.Error())
	}

	if len(allUsers) != 2 {
		t.Fatalf("Expected 2 users found: %d\n", len(allUsers))
	}

	if err := jdb.Delete("user", userA.ID); err != nil {
		t.Fatalf("Failed to delete user: %s\n", err.Error())
	}

	allUsers = []User{}

	if err := jdb.Select("user", &allUsers); err != nil {
		t.Fatalf("Failed to select all users: %s\n", err.Error())
	}

	if len(allUsers) != 1 {
		t.Fatalf("Expected 1 users found: %d\n", len(allUsers))
	}
}
