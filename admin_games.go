package main

import (
	"io/ioutil"
	"net/http"

	"encoding/base64"

	"github.com/gorilla/mux"
)

func handleAdminGames(w http.ResponseWriter, req *http.Request, page *pageData) {
	vars := mux.Vars(req)
	page.SubTitle = "Games"
	teamId := vars["id"]
	if teamId == "" {
		// Games List
		type gamesPageData struct {
			Teams []Team
		}
		gpd := new(gamesPageData)
		gpd.Teams = dbGetAllTeams()
		page.TemplateData = gpd
		page.SubTitle = "Games"
		page.show("admin-games.html", w)
	} else {
		switch vars["function"] {
		case "save":
			name := req.FormValue("gamename")
			desc := req.FormValue("gamedesc")
			if dbIsValidTeam(teamId) {
				if err := dbUpdateTeamGame(teamId, name, desc); err != nil {
					page.session.setFlashMessage("Error updating game: "+err.Error(), "error")
				} else {
					page.session.setFlashMessage("Team game updated", "success")
				}
				redirect("/admin/teams/"+teamId, w, req)
			}
		case "screenshotupload":
			if err := saveScreenshots(teamId, req); err != nil {
				page.session.setFlashMessage("Error updating game: "+err.Error(), "error")
			}
			redirect("/admin/teams/"+teamId, w, req)
		case "screenshotdelete":
			ssid := vars["subid"]
			if err := dbDeleteTeamGameScreenshot(teamId, ssid); err != nil {
				page.session.setFlashMessage("Error deleting screenshot: "+err.Error(), "error")
			}
			redirect("/admin/teams/"+teamId, w, req)
		}
	}
}

func saveScreenshots(teamId string, req *http.Request) error {
	file, _, err := req.FormFile("newssfile")
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(file)
	str := base64.StdEncoding.EncodeToString(data)
	return dbSaveTeamGameScreenshot(teamId, &Screenshot{Image: str})
}
