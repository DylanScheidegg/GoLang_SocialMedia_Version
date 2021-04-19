package accountcontroller

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
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

type showPost struct {
	FName string `json:"fName"`
	LName string `json:"lName"`
	Text  string `json:"Text"`
	Time  string `json:"Time"`
	Date  string `json:"Date"`
}

type FriendUser struct {
	ID    int    `json:"id"`
	FName string `json:"fName"`
	LName string `json:"lName"`
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
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

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

func LoadPosts(frList string) []showPost {
	// create the postgres db connection
	db := createConnection()

	// close the db connection
	defer db.Close()

	newList := strings.Split(frList, ",")

	var uPosts []showPost

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
			uPosts = append(uPosts, showPost{FName: "No Friends Have Posted Yet", LName: "", Text: "", Date: "", Time: ""})

		case nil:
			// create a User of models.User type
			user := User{}

			// execute the sql statement
			uRow := db.QueryRow(`SELECT * FROM users WHERE id=$1`, post.UID)

			// unmarshal the row object to user
			err2 := uRow.Scan(&user.Email, &user.FName, &user.ID, &user.LName, &user.Location, &user.Password, &user.Age, &user.Friends)

			switch err2 {
			case sql.ErrNoRows:
				uPosts = append(uPosts, showPost{FName: "No Friends Have Posted Yet", LName: "", Text: "", Date: "", Time: ""})
			case nil:
				uPosts = append(uPosts, showPost{FName: user.FName, LName: user.LName, Text: post.Text, Date: post.Date, Time: post.Time})
			}
		}
		count += 1
	}

	return uPosts
}

func LoadFriends(frList string) []FriendUser {
	// create the postgres db connection
	db := createConnection()

	// close the db connection
	defer db.Close()

	newList := strings.Split(frList, ",")

	var friends []FriendUser

	count := 0
	for _, fr := range newList {
		// create a posts of models.post type
		user := User{}

		// execute the sql statement
		row := db.QueryRow(`SELECT * FROM users WHERE id=$1`, fr)

		// unmarshal the row object to user
		err := row.Scan(&user.Email, &user.FName, &user.ID, &user.LName, &user.Location, &user.Password, &user.Age, &user.Friends)

		switch err {
		case sql.ErrNoRows:
			friends = append(friends, FriendUser{ID: -1, FName: "No Friends Found", LName: ""})
		case nil:
			friends = append(friends, FriendUser{ID: user.ID, FName: user.FName, LName: user.LName})
		}
		count += 1
	}
	//fmt.Println(friends)
	return friends
}

func Logout(response http.ResponseWriter, request *http.Request) {
	session, _ := store.Get(request, "mysession")
	session.Options.MaxAge = -1
	session.Save(request, response)
	http.Redirect(response, request, "/", http.StatusSeeOther)
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
	hash := sha256.New()
	hash.Write([]byte(request.Form.Get("password")))
	passwordHashed := hex.EncodeToString(hash.Sum(nil))

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

		tmp, _ := template.ParseFiles("views/accountcontroller/home.html")
		tmp.Execute(response, data)
	}
}

func FetchPosts(response http.ResponseWriter, request *http.Request) {
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

	if err != nil {
		fmt.Println("Error loading posts")
	}

	posts := LoadPosts(user.Friends)

	postsListBytes, errMar := json.Marshal(posts)

	//fmt.Println(posts[0].Text)
	if errMar != nil {
		fmt.Println("Error: Loading Posts")
	}

	response.Write(postsListBytes)
}

func FetchFriends(response http.ResponseWriter, request *http.Request) {
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

	if err != nil {
		fmt.Println("Error loading friends")
	}

	friends := LoadFriends(user.Friends)
	//fmt.Println(friends[0].FName)

	friendsListBytes, errMar := json.Marshal(friends)

	//fmt.Println(posts[0].Text)
	if errMar != nil {
		fmt.Println("Error: Loading friends")
	}

	response.Write(friendsListBytes)
}
