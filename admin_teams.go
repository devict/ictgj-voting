package main

import (
	"net/http"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/gorilla/mux"
)

func refreshTeamsInMemory() {
	site.Teams = db.getAllTeams()
}

func handleAdminTeams(w http.ResponseWriter, req *http.Request, page *pageData) {
	vars := mux.Vars(req)
	page.SubTitle = "Teams"
	teamId := vars["id"]
	if teamId == "new" {
		// Add a new team
		switch vars["function"] {
		case "save":
			name := req.FormValue("teamname")
			if db.getTeamByName(name) != nil {
				// A team with that name already exists
				page.session.setFlashMessage("A team with the name "+name+" already exists!", "error")
			} else {
				if err := db.newTeam(name); err != nil {
					page.session.setFlashMessage(err.Error(), "error")
				} else {
					page.session.setFlashMessage("Team "+name+" created!", "success")
				}
			}
			refreshTeamsInMemory()
			redirect("/admin/teams", w, req)
		default:
			page.SubTitle = "Add New Team"
			page.show("admin-addteam.html", w)
		}
	} else if teamId != "" {
		// Functions for existing team
		tm := db.getTeam(teamId)
		if tm != nil {
			switch vars["function"] {
			case "save":
				tm.UUID = teamId
				tm.Name = req.FormValue("teamname")
				if err := tm.save(); err != nil {
					page.session.setFlashMessage("Error updating team: "+err.Error(), "error")
				} else {
					page.session.setFlashMessage("Team Updated!", "success")
				}
				refreshTeamsInMemory()
				redirect("/admin/teams", w, req)
			case "delete":
				var err error
				if err = tm.delete(); err != nil {
					page.session.setFlashMessage("Error deleting team: "+err.Error(), "error")
				} else {
					page.session.setFlashMessage("Team "+tm.Name+" Deleted", "success")
				}
				refreshTeamsInMemory()
				redirect("/admin/teams", w, req)
			case "savemember":
				mbrName := req.FormValue("newmembername")
				mbr := newTeamMember(mbrName)
				mbr.SlackId = req.FormValue("newmemberslackid")
				mbr.Twitter = req.FormValue("newmembertwitter")
				mbr.Email = req.FormValue("newmemberemail")
				if err := tm.updateTeamMember(mbr); err != nil {
					page.session.setFlashMessage("Error adding team member: "+err.Error(), "error")
				} else {
					page.session.setFlashMessage(mbrName+" added to team!", "success")
				}
				refreshTeamsInMemory()
				redirect("/admin/teams/"+teamId+"#members", w, req)
			case "deletemember":
				m := tm.getTeamMember(req.FormValue("memberid"))
				if m != nil {
					if err := tm.deleteTeamMember(m); err != nil {
						page.session.setFlashMessage("Error deleting team member: "+err.Error(), "error")
					} else {
						page.session.setFlashMessage(m.Name+" deleted from team", "success")
					}
					refreshTeamsInMemory()
				} else {
					page.session.setFlashMessage("Couldn't find team member to delete", "error")
				}
				redirect("/admin/teams/"+teamId+"#members", w, req)
			default:
				page.SubTitle = "Edit Team"
				t := db.getTeam(teamId)
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
		page.TemplateData = teamsPageData{Teams: db.getAllTeams()}
		page.SubTitle = "Teams"
		page.show("admin-teams.html", w)
	}
}
