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
		// Add a new team
		switch vars["function"] {
		case "save":
			name := req.FormValue("teamname")
			if dbGetTeamByName(name) != nil {
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
		// Functions for existing team
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
				t := dbGetTeam(teamId)
				if err = dbDeleteTeam(teamId); err != nil {
					page.session.setFlashMessage("Error deleting team: "+err.Error(), "error")
				} else {
					page.session.setFlashMessage("Team "+t.Name+" Deleted", "success")
				}
				redirect("/admin/teams", w, req)
			case "savemember":
				mbrName := req.FormValue("newmembername")
				mbrSlack := req.FormValue("newmemberslackid")
				mbrTwitter := req.FormValue("newmembertwitter")
				mbrEmail := req.FormValue("newmemberemail")
				if err := dbAddTeamMember(teamId, mbrName, mbrEmail, mbrSlack, mbrTwitter); err != nil {
					page.session.setFlashMessage("Error adding team member: "+err.Error(), "error")
				} else {
					page.session.setFlashMessage(mbrName+" added to team!", "success")
				}
				redirect("/admin/teams/"+teamId, w, req)
			case "deletemember":
				mbrId := req.FormValue("memberid")
				m, _ := dbGetTeamMember(teamId, mbrId)
				if err := dbDeleteTeamMember(teamId, mbrId); err != nil {
					page.session.setFlashMessage("Error deleting team member: "+err.Error(), "error")
				} else {
					page.session.setFlashMessage(m.Name+" deleted from team", "success")
				}
				redirect("/admin/teams/"+teamId, w, req)
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
		// Team List
		type teamsPageData struct {
			Teams []Team
		}
		page.TemplateData = teamsPageData{Teams: dbGetAllTeams()}
		page.SubTitle = "Teams"
		page.show("admin-teams.html", w)
	}
}
