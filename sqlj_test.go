package sqlj

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID        uint      `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
}

func TestInsertAndRetrieve(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")

	defer db.Close()

	db.Exec("CREATE TABLE user (id integer primary key, name text, email text, created_at timestamp)")

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

func TestTransaction(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")

	defer db.Close()

	db.Exec("CREATE TABLE user (id integer primary key, name text, email text, created_at timestamp)")

	if err != nil {
		t.Fatalf("Failed to open db: %s\n", err.Error())
	}

	tx, err := db.Begin()

	if err != nil {
		t.Fatalf("Failed to start transaction: %s\n", err.Error())
	}

	jdb := DB{
		DB:           tx,
		SkipOnInsert: []string{"id"},
	}

	user := User{
		Name:  "Jess",
		Email: "jess@example.com",
	}

	if err := jdb.Insert("user", &user); err != nil {
		t.Fatalf("Failed to insert user: %s\n", err.Error())
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit: %s\n", err.Error())
	}

	jdb = DB{
		DB: db,
	}

	foundUser := User{}

	if err := jdb.Get("user", user.ID, &foundUser); err != nil {
		t.Fatalf("Failed to retrieve user: %s\n", err.Error())
	}
}

func TestInsertWithOptions(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")

	defer db.Close()

	db.Exec("CREATE TABLE user (id integer primary key, name text, email text, created_at timestamp)")

	if err != nil {
		t.Fatalf("Failed to open db: %s\n", err.Error())
	}

	jdb := DB{
		DB:           db,
		SkipOnInsert: []string{"id"},
	}

	user := User{
		Name:  "Jess",
		Email: "jess@example.com",
	}

	if err := jdb.InsertWithOptions("user", Options{
		Fields: []Field{
			BasicField{
				Name:  "name",
				Value: "Jon",
			},
		},
	}, &user); err != nil {
		t.Fatalf("Failed to insert user: %s\n", err.Error())
	}

	if user.Name != "Jon" {
		t.Fatalf("Expected user.Name to \"Jon\". Got: %s\n", user.Name)
	}

	foundUser := User{}

	if err := jdb.Get("user", user.ID, &foundUser); err != nil {
		t.Fatalf("Failed to retrieve user: %s\n", err.Error())
	}

	if err := jdb.InsertWithOptions("user", Options{
		Fields: []Field{
			LiteralField{
				Name:  "created_at",
				Value: "date()",
			},
		},
	}, &user); err != nil {
		t.Fatalf("Failed to insert user: %s\n", err.Error())
	}

	if err := jdb.Get("user", user.ID, &foundUser); err != nil {
		t.Fatalf("Failed to retrieve user: %s\n", err.Error())
	}
}

func TestFrom(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")

	defer db.Close()

	db.Exec("CREATE TABLE user (id integer primary key, name text, email text, created_at timestamp)")

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
	if err := jdb.From("user").Select(&allUsers); err != nil {
		t.Fatalf("Fluent API, failed to select all: %s\n", err.Error())
	}

	if len(allUsers) != 3 {
		t.Fatalf("Fluent API, did not retrieve all users. Expected 3, Got: %d\n", len(allUsers))
	}

	var sortedUsers []User
	if err := jdb.From("user").Order("name", "ASC").Select(&sortedUsers); err != nil {
		t.Fatalf("Fluent API, failed to select all: %s\n", err.Error())
	}

	if sortedUsers[0].Name != "Jane" {
		t.Fatalf("Fluent API, sorting failed")
	}

	var reverseSortedUsers []User
	if err := jdb.From("user").Order("name", "DESC").Select(&reverseSortedUsers); err != nil {
		t.Fatalf("Fluent API, failed to select all: %s\n", err.Error())
	}

	if reverseSortedUsers[0].Name != "Joe" {
		t.Fatalf("Fluent API, sorting failed")
	}
}
