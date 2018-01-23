package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func handleAdminArchive(w http.ResponseWriter, req *http.Request, page *pageData) {
	vars := mux.Vars(req)
	page.SubTitle = "GameJam Archive"
	id := vars["id"]
	if id == "" {
		// Archive List
		type archivePageData struct {
			Gamejams []Gamejam
		}
		//apd := new(archivePageData)
	}
}
