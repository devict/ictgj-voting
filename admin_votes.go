package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func handleAdminVotes(w http.ResponseWriter, req *http.Request, page *pageData) {
	vars := mux.Vars(req)
	page.SubTitle = "Votes"
	switch vars["function"] {
	default:
		type votesPageData struct {
			Votes []Vote
		}
		page.TemplateData = votesPageData{Votes: dbGetAllVotes()}
		page.show("admin-votes.html", w)
	}
}
