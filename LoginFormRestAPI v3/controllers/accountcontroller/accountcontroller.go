package accountcontroller

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/sessions"

	_ "github.com/lib/pq"
)

var store = sessions.NewCookieStore([]byte("mysession"))

type User struct {
	Email    string `json:"email"`
	FName    string `json:"fname"`
	ID       int    `json:"id"`
	LName    string `json:"lname"`
	Location string `json:"location"`
	Password string `json:"password"`
	Age      int    `json:"age"`
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "PASSWORD HERE"
	dbname   = "socialMedia"
)

// create connection with postgres db
func createConnection() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Open the connection
	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		panic(err)
	}

	// check the connection
	err = db.Ping()

	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
	// return the connection
	return db
}

func LoginPage(response http.ResponseWriter, request *http.Request) {
	tmp, _ := template.ParseFiles("views/accountcontroller/login.html")
	tmp.Execute(response, nil)
}

func Login(response http.ResponseWriter, request *http.Request) {
	// create the postgres db connection
	db := createConnection()

	// close the db connection
	defer db.Close()

	// create a user of models.User type
	user := User{}

	request.ParseForm()
	email := request.Form.Get("email")
	fmt.Println(email)

	// SHA 256 Password
	h := sha256.New()
	h.Write([]byte(request.Form.Get("password")))
	passwordHashed := hex.EncodeToString(h.Sum(nil))

	// execute the sql statement
	row := db.QueryRow(`SELECT * FROM users WHERE email=$1 AND password=$2`, email, passwordHashed)

	// unmarshal the row object to user
	err := row.Scan(&user.Email, &user.FName, &user.ID, &user.LName, &user.Location, &user.Password, &user.Age)

	// Checking if user exists
	switch err {
	case sql.ErrNoRows:
		data := map[string]interface{}{
			"err": "Inavlid",
		}
		tmp, _ := template.ParseFiles("views/accountcontroller/login.html")
		tmp.Execute(response, data)
	case nil:
		fmt.Println("hi")
		session, _ := store.Get(request, "mysession")
		session.Values["id"] = user.ID
		session.Values["email"] = user.Email
		session.Save(request, response)
		http.Redirect(response, request, "/home", http.StatusSeeOther)
	}
}

func RegisterPage(response http.ResponseWriter, request *http.Request) {
	tmp, _ := template.ParseFiles("views/accountcontroller/register.html")
	tmp.Execute(response, nil)
}

func Register(response http.ResponseWriter, request *http.Request) {
	// create the postgres db connection
	db := createConnection()

	// close the db connection
	defer db.Close()

	// create a user of models.User type
	user := User{}

	request.ParseForm()
	email := request.Form.Get("email")
	fname := request.Form.Get("fname")
	lname := request.Form.Get("lname")
	location := request.Form.Get("location")
	age := request.Form.Get("age")

	//SHA 256 Password
	h := sha256.New()
	h.Write([]byte(request.Form.Get("password")))
	passwordHashed := hex.EncodeToString(h.Sum(nil))

	checkRow := db.QueryRow(`SELECT * FROM users WHERE email=$1`, email)

	// unmarshal the row object to user
	err := checkRow.Scan(&user.Email, &user.FName, &user.ID, &user.LName, &user.Location, &user.Password, &user.Age)

	// Checking if user exists
	switch err {
	case sql.ErrNoRows: // NO USER FOUND WITH GIVEN EMAIL
		var id int
		db.QueryRow(`SELECT COUNT(*) users`).Scan(&id)

		// execute the sql statement to INSERT
		err2 := db.QueryRow(`INSERT INTO users (id, email, fname, lname, location, password, age) VALUES ($1, $2, $3, $4, $5, $6, $7)`, (id + 1), email, fname, lname, location, passwordHashed, age)

		if err2 != nil {
			log.Fatalf("Unable to execute the query. %v", err2)
		}

		fmt.Printf("Inserted a single record %v", id)

		http.Redirect(response, request, "/", http.StatusSeeOther)
	case nil:
		data := map[string]interface{}{
			"err": "Inavlid Email Already In Use",
		}
		tmp, _ := template.ParseFiles("views/accountcontroller/register.html")
		tmp.Execute(response, data)
	default:
		data := map[string]interface{}{
			"err": "Inavlid",
		}
		tmp, _ := template.ParseFiles("views/accountcontroller/register.html")
		tmp.Execute(response, data)
	}
}

func Home(response http.ResponseWriter, request *http.Request) {
	// create the postgres db connection
	db := createConnection()

	// close the db connection
	defer db.Close()

	// create a user of models.User type
	user := User{}

	session, _ := store.Get(request, "mysession")
	id := session.Values["id"]
	ema := session.Values["email"]

	// execute the sql statement
	row := db.QueryRow(`SELECT * FROM users WHERE id=$1 AND email=$2`, id, ema)

	// unmarshal the row object to user
	err := row.Scan(&user.Email, &user.FName, &user.ID, &user.LName, &user.Location, &user.Age, &user.Password)

	// Checking if user exists
	switch err {
	case sql.ErrNoRows:
		data := map[string]interface{}{
			"err": "No Row Returned",
		}
		tmp, _ := template.ParseFiles("views/accountcontroller/login.html")
		tmp.Execute(response, data)
	default:
		data := map[string]interface{}{
			"id":       user.ID,
			"fname":    user.FName,
			"lname":    user.LName,
			"email":    user.Email,
			"location": user.Location,
			"age":      user.Age,
		}

		tmp, _ := template.ParseFiles("views/accountcontroller/home.html")
		tmp.Execute(response, data)
	}
}

func Logout(response http.ResponseWriter, request *http.Request) {
	session, _ := store.Get(request, "mysession")
	session.Options.MaxAge = -1
	session.Save(request, response)
	http.Redirect(response, request, "/", http.StatusSeeOther)
}
