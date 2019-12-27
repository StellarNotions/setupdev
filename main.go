package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type notFoundHttpResponse struct {
	Msg string `json:"msg"`
}

func main() {
	loadedSuccessfully := loadENVs(ENVFilePath, ENVFileName)
	if !loadedSuccessfully {
		log.Fatal("Unable to load ENV file")
	}

	router := mux.NewRouter()
	router.HandleFunc("/", homeLink)
	router.HandleFunc("/{name}/{version}/{operatingSystem}", getSetupPackage)

	fmt.Print("Running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
