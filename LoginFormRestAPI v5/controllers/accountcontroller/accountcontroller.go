package accountcontroller

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
	_ "time"

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
	Friends  string `json:"friends"`
}

type Post struct {
	ID   int    `json:"id"`
	UID  int    `json:"uid"`
	Text string `json:"text"`
	Date string `json:"date"`
	Time string `json:"time"`
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

func LoadPosts(frList string) [2][5]string {
	// create the postgres db connection
	db := createConnection()

	// close the db connection
	defer db.Close()

	newList := strings.Split(frList, ",")

	//posts := [len(newList)][5]string{}
	posts := [2][5]string{}

	count := 0
	for _, fr := range newList {
		// create a posts of models.post type
		post := Post{}

		// execute the sql statement
		row := db.QueryRow(`SELECT * FROM posts WHERE uID=$1`, fr)

		// unmarshal the row object to user
		err := row.Scan(&post.ID, &post.UID, &post.Text, &post.Date, &post.Time)

		switch err {
		case sql.ErrNoRows:
			posts[count][0] = "No Posts Available"
			posts[count][1] = ""
			posts[count][2] = ""
			posts[count][3] = ""
			posts[count][4] = ""
		case nil:
			// create a User of models.User type
			user := User{}

			// execute the sql statement
			uRow := db.QueryRow(`SELECT * FROM users WHERE id=$1`, post.UID)

			// unmarshal the row object to user
			err2 := uRow.Scan(&user.Email, &user.FName, &user.ID, &user.LName, &user.Location, &user.Password, &user.Age, &user.Friends)

			fmt.Println(err2)
			switch err2 {
			case sql.ErrNoRows:
				posts[count][0] = "No Posts Available"
				posts[count][1] = ""
				posts[count][2] = ""
				posts[count][3] = ""
				posts[count][4] = ""
			case nil:
				posts[count][0] = user.FName
				posts[count][1] = user.LName
				posts[count][2] = post.Text
				posts[count][3] = post.Date
				posts[count][4] = post.Time
			}
		}
		count += 1
	}
	return posts
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

	// SHA 256 Password
	h := sha256.New()
	h.Write([]byte(request.Form.Get("password")))
	passwordHashed := hex.EncodeToString(h.Sum(nil))

	// execute the sql statement
	row := db.QueryRow(`SELECT * FROM users WHERE email=$1 AND password=$2`, email, passwordHashed)

	// unmarshal the row object to user
	err := row.Scan(&user.Email, &user.FName, &user.ID, &user.LName, &user.Location, &user.Password, &user.Age, &user.Friends)

	// Checking if user exists
	switch err {
	case sql.ErrNoRows:
		data := map[string]interface{}{
			"err": "Inavlid",
		}
		tmp, _ := template.ParseFiles("views/accountcontroller/login.html")
		tmp.Execute(response, data)
	case nil:
		session, _ := store.Get(request, "mysession")
		session.Values["id"] = user.ID
		session.Values["email"] = user.Email
		session.Save(request, response)
		http.Redirect(response, request, "/home", http.StatusSeeOther)
	default:
		tmp, _ := template.ParseFiles("views/accountcontroller/login.html")
		tmp.Execute(response, nil)
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
	err := checkRow.Scan(&user.Email, &user.FName, &user.ID, &user.LName, &user.Location, &user.Password, &user.Age, &user.Friends)

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
		tmp, _ := template.ParseFiles("views/accountcontroller/register.html")
		tmp.Execute(response, nil)
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
	err := row.Scan(&user.Email, &user.FName, &user.ID, &user.LName, &user.Location, &user.Password, &user.Age, &user.Friends)

	// Checking if user exists
	switch err {
	case sql.ErrNoRows:
		data := map[string]interface{}{
			"err": "No Row Returned",
		}
		tmp, _ := template.ParseFiles("views/accountcontroller/login.html")
		tmp.Execute(response, data)
	case nil:
		data := map[string]interface{}{
			"id":       user.ID,
			"fname":    user.FName,
			"lname":    user.LName,
			"email":    user.Email,
			"location": user.Location,
			"age":      user.Age,
			"friends":  user.Friends,
		}

		fmt.Println(LoadPosts(user.Friends))

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
