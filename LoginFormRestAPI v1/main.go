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

	router.HandleFunc("/account", accountcontroller.Index)
	router.HandleFunc("/account/index", accountcontroller.Index)
	router.HandleFunc("/account/login", accountcontroller.Login)
	router.HandleFunc("/account/welcome", accountcontroller.Welcome)
	router.HandleFunc("/account/logout", accountcontroller.Logout)

	return router
}

//https://www.youtube.com/watch?v=CROR9tWLgFo
//http://localhost:10000/account

func main() {
	r := Router()

	fmt.Println("Starting server on the port 10000...")
	log.Fatal(http.ListenAndServe(":10000", r))
}
