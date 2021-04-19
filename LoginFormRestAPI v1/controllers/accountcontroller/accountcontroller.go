package accountcontroller

import (
	"database/sql"
	"fmt"
	"net/http"
	"text/template"

	"github.com/gorilla/sessions"

	_ "github.com/lib/pq"
)

var store = sessions.NewCookieStore([]byte("mysession"))

type User struct {
	Email    string `json:"email"`
	FName    string `json:"fname"`
	ID       int64  `json:"id"`
	LName    string `json:"lname"`
	Location string `json:"location"`
	Password string `json:"password"`
	Age      int64  `json:"age"`
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

func Index(response http.ResponseWriter, request *http.Request) {
	tmp, _ := template.ParseFiles("views/accountcontroller/index.html")
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
	password := request.Form.Get("password")
	fmt.Println(email, password)

	// execute the sql statement
	row := db.QueryRow(`SELECT * FROM users WHERE email=$1 AND password=$2`, email, password)

	// unmarshal the row object to user
	err := row.Scan(&user.Email, &user.FName, &user.ID, &user.LName, &user.Location, &user.Password, &user.Age)

	switch err {
	case sql.ErrNoRows:
		data := map[string]interface{}{
			"err": "Inavlid",
		}
		tmp, _ := template.ParseFiles("views/accountcontroller/index.html")
		tmp.Execute(response, data)
	case nil:
		session, _ := store.Get(request, "mysession")
		session.Values["email"] = email
		session.Save(request, response)
		http.Redirect(response, request, "/account/welcome", http.StatusSeeOther)
	default:
		data := map[string]interface{}{
			"err": "Inavlid",
		}
		tmp, _ := template.ParseFiles("views/accountcontroller/index.html")
		tmp.Execute(response, data)
	}
}

func Welcome(response http.ResponseWriter, request *http.Request) {
	session, _ := store.Get(request, "mysession")
	email := session.Values["email"]
	data := map[string]interface{}{
		"email": email,
	}

	tmp, _ := template.ParseFiles("views/accountcontroller/welcome.html")
	tmp.Execute(response, data)
}

func Logout(response http.ResponseWriter, request *http.Request) {
	session, _ := store.Get(request, "mysession")
	session.Options.MaxAge = -1
	session.Save(request, response)
	http.Redirect(response, request, "/account/index", http.StatusSeeOther)
}
