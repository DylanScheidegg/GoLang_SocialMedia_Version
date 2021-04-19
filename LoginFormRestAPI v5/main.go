package main

import (
	"fmt"
	"log"
	"net/http"

	"./controllers/accountcontroller"
	"github.com/gorilla/mux"
)

func Router() *mux.Router {
	//fs := http.FileServer(http.Dir("./static"))
	//http.Handle("/", fs)
	router := mux.NewRouter()

	router.HandleFunc("/", accountcontroller.LoginPage)
	router.HandleFunc("/login", accountcontroller.Login)
	router.HandleFunc("/logout", accountcontroller.Logout)
	router.HandleFunc("/registerpage", accountcontroller.RegisterPage)
	router.HandleFunc("/register", accountcontroller.Register)
	router.HandleFunc("/home", accountcontroller.Home)
	router.HandleFunc("/logout", accountcontroller.Logout)

	return router
}

//https://www.youtube.com/watch?v=CROR9tWLgFo
//http://localhost:10000/account

func main() {
	r := Router()

	fmt.Println("Starting server on the port 10000...")
	log.Fatal(http.ListenAndServe(":10000", r))
}
