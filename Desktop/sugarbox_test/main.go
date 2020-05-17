package main

import (
    "fmt"
    "os"
	"net/http"

	"github.com/gorilla/mux"

	"./app"
)

func main () {

	router := mux.NewRouter()
	router.HandleFunc("/api/rating/new/{mname}", app.AddUserRating).Methods("POST")
	router.HandleFunc("/api/comment/new/{mname}", app.AddUserComment).Methods("POST")
	router.HandleFunc("/api/movies", app.FindAllMovies).Methods("GET")
	router.HandleFunc("/api/user_activity", app.FetchUserActivity).Methods("GET")

	port := os.Getenv("PORT") //Get port from .env file, we did not specify any port so this should return an empty string when tested locally
	if port == "" {
		port = "8000" //localhost
	}

	fmt.Println(port)

	err := http.ListenAndServe(":" + port, router) //Launch the app, visit localhost:8000/api
	if err != nil {
		fmt.Print(err)
	}


}
