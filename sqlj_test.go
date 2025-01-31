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

	userB.Email = "jack@example.com"

	if err := jdb.Update("user", userB.ID, &userB); err != nil {
		t.Fatalf("Failed to update user: %s\n", err.Error())
	}

	jackUser := User{}

	if err := jdb.Get("user", userB.ID, &jackUser); err != nil {
		t.Fatalf("Failed to get updated user: %s\n", err.Error())
	}

	if jackUser.Email != "jack@example.com" {
		t.Fatalf("User email did not update")
	}
}

type Employee struct {
	ID        uint   `db:"id"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Email     string `db:"email"`
	Location  string `db:"location"`
	Age       uint   `db:"age"`
}

type EmailAndLocation struct {
	Email    string `db:"email"`
	Location string `db:"location"`
}

type EmployeeShort struct {
	ID    uint   `db:"id"`
	Email string `db:"email"`
}

func TestPartialRetrieval(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")

	defer db.Close()

	db.Exec("CREATE TABLE employee (id integer primary key, first_name text, last_name text, email text, location text, age integer)")

	if err != nil {
		t.Fatalf("Failed to open db: %s\n", err.Error())
	}

	jdb := DB{
		DB:           db,
		SkipOnInsert: []string{"id"},
	}

	employeeA := Employee{
		FirstName: "Joe",
		LastName:  "Smith",
		Email:     "joe@example.com",
		Location:  "England",
		Age:       21,
	}
	employeeB := Employee{
		FirstName: "Jen",
		LastName:  "Jones",
		Email:     "jen@example.com",
		Location:  "Wales",
		Age:       28,
	}

	if err := jdb.Insert("employee", &employeeA); err != nil {
		t.Fatalf("Failed to insert employee: %s\n", err.Error())
	}

	if err := jdb.Insert("employee", &employeeB); err != nil {
		t.Fatalf("Failed to insert employee: %s\n", err.Error())
	}

	emailLocA := EmailAndLocation{}

	if err := jdb.Get("employee", employeeA.ID, &emailLocA); err != nil {
		t.Fatalf("Failed to get partial employee: %s\n", err.Error())
	}

	if emailLocA.Email != employeeA.Email || emailLocA.Location != employeeA.Location {
		t.Fatal("Data mismatch partial employee\n")
	}

	emailLocs := []EmailAndLocation{}

	if err := jdb.Select("employee", &emailLocs); err != nil {
		t.Fatalf("Failed to select partial employees: %s\n", err.Error())
	}

	shortEmployee := EmployeeShort{
		Email: "jon@example.com",
	}

	if err := jdb.Insert("employee", &shortEmployee); err != nil {
		t.Fatalf("Failed to insert short employee: %s\n", err.Error())
	}
}
