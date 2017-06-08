package main

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func handleAdminGames(w http.ResponseWriter, req *http.Request, page *pageData) {
	vars := mux.Vars(req)
	page.SubTitle = "Games"
	gameId := vars["id"]
	teamId := req.FormValue("teamid")
	if strings.TrimSpace(teamId) != "" {
		page.session.setStringValue("teamid", teamId)
		page.TeamID = teamId
	}
	if gameId == "new" {
		switch vars["function"] {
		case "save":
			name := req.FormValue("gamename")
			if dbIsValidTeam(teamId) {
				if dbEditTeamGame(teamId, name) != nil {
				}
			}
		default:
			page.SubTitle = "Add New Game"
			page.show("admin-addgame.html", w)
		}
	}
}
