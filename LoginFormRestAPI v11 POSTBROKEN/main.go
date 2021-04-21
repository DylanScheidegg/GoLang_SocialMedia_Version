package main

import (
	"fmt"
	"log"
	"net/http"

	"./controllers/accountcontroller"
	"github.com/gorilla/mux"
)

func Router() *mux.Router {
	router := mux.NewRouter()

	// Pages
	router.HandleFunc("/", accountcontroller.LoginPage)
	router.HandleFunc("/register", accountcontroller.Register)
	router.HandleFunc("/home", accountcontroller.Home)

	// Handlers
	router.HandleFunc("/postPost", accountcontroller.PostPost)
	router.HandleFunc("/fetchPosts", accountcontroller.FetchPosts)
	router.HandleFunc("/fetchFriends", accountcontroller.FetchFriends)
	router.HandleFunc("/login", accountcontroller.Login)
	router.HandleFunc("/registerpage", accountcontroller.RegisterPage)
	router.HandleFunc("/logout", accountcontroller.Logout)

	return router
}

func main() {
	r := Router()

	fmt.Println("Starting server on the port 10000...")
	log.Fatal(http.ListenAndServe(":10000", r))
}
