package sqlj

import (
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID        uint      `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
}

var pgDB *sql.DB

func TestMain(m *testing.M) {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("No .env file found")
	}

	var err error
	pgDB, err = sql.Open("postgres", os.Getenv("PG_DSN"))

	if err != nil {
		log.Fatalf("Failed to open pg db: ")
	}

	defer pgDB.Close()

	m.Run()
}

func TestOpenAndClose(t *testing.T) {
	jdb, err := Open("sqlite3", ":memory:")

	if err != nil {
		t.Fatalf("Failed to open DB: %s\n", err.Error())
	}

	defer jdb.Close()

	row := jdb.DB.QueryRow("SELECT 1")

	var val uint

	if err := row.Scan(&val); err != nil {
		t.Fatalf("Failed to scan val: %s\n", err.Error())
	}

	if val != 1 {
		t.Fatalf("Expected value to be 1, got: %d\n", val)
	}
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

	jdb := NewDB(db)

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
	UserID    *uint  `db:"user_id"`
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

	db.Exec("CREATE TABLE employee (id integer primary key, first_name text, last_name text, email text, location text, age integer, user_id integer)")

	if err != nil {
		t.Fatalf("Failed to open db: %s\n", err.Error())
	}

	jdb := NewDB(db)

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

	jdb := NewDB(tx)

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

func TestInsertWithFields(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")

	defer db.Close()

	db.Exec("CREATE TABLE user (id integer primary key, name text, email text, created_at timestamp)")

	if err != nil {
		t.Fatalf("Failed to open db: %s\n", err.Error())
	}

	jdb := NewDB(db)

	user := User{
		Name:  "Jess",
		Email: "jess@example.com",
	}

	// This is probably a good example of what not to do
	// anything that isn't a literal value should come from v struct
	if err := jdb.InsertWithFields("user", &user, map[string]string{"name": "'Jon'"}); err != nil {
		t.Fatalf("Failed to insert user: %s\n", err.Error())
	}

	if user.Name != "Jon" {
		t.Fatalf("Expected user.Name to \"Jon\". Got: %s\n", user.Name)
	}

	foundUser := User{}

	if err := jdb.Get("user", user.ID, &foundUser); err != nil {
		t.Fatalf("Failed to retrieve user: %s\n", err.Error())
	}

	if err := jdb.InsertWithFields("user", &user, map[string]string{"created_at": "date()"}); err != nil {
		t.Fatalf("Failed to insert user: %s\n", err.Error())
	}

	if err := jdb.Get("user", user.ID, &foundUser); err != nil {
		t.Fatalf("Failed to retrieve user: %s\n", err.Error())
	}
}

type Issue struct {
	ID         uint   `db:"id"`
	Title      string `db:"title"`
	NotInDB    string
	AssignedTo *uint `db:"assigned_to"`
}

func TestNullableFields(t *testing.T) {
	if _, err := pgDB.Exec("CREATE TABLE issues (id serial primary key, title text, assigned_to integer)"); err != nil {
		t.Fatalf("Failed to set-up postgres DB: %s\n", err.Error())
	}

	t.Cleanup(func() {
		if _, err := pgDB.Exec("DROP TABLE issues"); err != nil {
			t.Logf("Cleanup - Failed to remove pg issues table: %s\n", err.Error())
		}
	})

	jdb := NewDB(pgDB)

	issueA := Issue{
		Title:      "A first issue",
		AssignedTo: nil,
	}

	issueB := Issue{
		Title:      "A second issue",
		AssignedTo: nil,
	}

	if err := jdb.Insert("issues", &issueA); err != nil {
		t.Fatalf("Failed to insert issue: %s\n", err.Error())
	}

	if err := jdb.Insert("issues", &issueB); err != nil {
		t.Fatalf("Failed to insert issue: %s\n", err.Error())
	}

	assignee := uint(1)

	issueA.AssignedTo = &assignee

	if err := jdb.Update("issues", issueA.ID, &issueA); err != nil {
		t.Fatalf("Failed to updated issue: %s\n", err.Error())
	}

	issueQuery := jdb.From("issues").Order("title", "ASC")

	var foundIssue Issue

	if err := issueQuery.Get(issueA.ID, &foundIssue); err != nil {
		t.Fatalf("Failed to get issue by ID: %s\n", err.Error())
	}

	if foundIssue.Title != issueA.Title {
		t.Fatalf("Title mismatch, expected: %s, got: %s\n", issueA.Title, foundIssue.Title)
	}

	var allIssues []Issue

	if err := issueQuery.All(&allIssues); err != nil {
		t.Fatalf("Failed to get all issues: %s\n", err.Error())
	}

	if len(allIssues) != 2 {
		t.Fatalf("Expected to get 2 issues. Got: %d issue(s)\n", len(allIssues))
	}

	var pageOne []Issue

	if err := issueQuery.Page(1, 10, &pageOne); err != nil {
		t.Fatalf("Failed to page issues: %s\n", err.Error())
	}
}
