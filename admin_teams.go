package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func handleAdminTeams(w http.ResponseWriter, req *http.Request, page *pageData) {
	vars := mux.Vars(req)
	page.SubTitle = "Teams"
	teamId := vars["id"]
	if teamId == "new" {
		switch vars["function"] {
		case "save":
			name := req.FormValue("teamname")
			if dbIsValidTeam(name) {
				// A team with that name already exists
				page.session.setFlashMessage("A team with the name "+name+" already exists!", "error")
			} else {
				if err := dbCreateNewTeam(name); err != nil {
					page.session.setFlashMessage(err.Error(), "error")
				} else {
					page.session.setFlashMessage("Team "+name+" created!", "success")
				}
			}
			redirect("/admin/teams", w, req)
		default:
			page.SubTitle = "Add New Team"
			page.show("admin-addteam.html", w)
		}
	} else if teamId != "" {
		if dbIsValidTeam(teamId) {
			switch vars["function"] {
			case "save":
				tm := new(Team)
				tm.UUID = teamId
				tm.Name = req.FormValue("teamname")
				if err := dbUpdateTeam(teamId, tm); err != nil {
					page.session.setFlashMessage("Error updating team: "+err.Error(), "error")
				} else {
					page.session.setFlashMessage("Team Updated!", "success")
				}
				redirect("/admin/teams", w, req)
			case "delete":
				var err error
				if err = dbDeleteTeam(teamId); err != nil {
					page.session.setFlashMessage("Error deleting team: "+err.Error(), "error")
				}
				redirect("/admin/teams", w, req)
			default:
				page.SubTitle = "Edit Team"
				t := dbGetTeam(teamId)
				page.TemplateData = t
				page.show("admin-editteam.html", w)
			}
		} else {
			page.session.setFlashMessage("Couldn't find the requested team, please try again.", "error")
			redirect("/admin/teams", w, req)
		}
	} else {
		type teamsPageData struct {
			Teams []Team
		}

		page.TemplateData = teamsPageData{Teams: dbGetAllTeams()}
		page.SubTitle = "Teams"
		page.show("admin-teams.html", w)
	}
}
