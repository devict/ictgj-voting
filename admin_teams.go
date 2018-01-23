package main

import (
	"fmt"
	"net/http"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

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
			tm := NewTeam("")
			tm.Name = name
			if err := m.jam.AddTeam(tm); err != nil {
				page.session.setFlashMessage("Error adding team: "+err.Error(), "error")
			}
			redirect("/admin/teams", w, req)
		default:
			page.SubTitle = "Add New Team"
			page.show("admin-addteam.html", w)
		}
	} else if teamId != "" {
		// Functions for existing team
		tm, _ := m.jam.GetTeamById(teamId)
		if tm != nil {
			switch vars["function"] {
			case "save":
				tm.Name = req.FormValue("teamname")
				page.session.setFlashMessage("Team Updated!", "success")
				redirect("/admin/teams", w, req)
			case "delete":
				var err error
				if err = m.jam.RemoveTeamById(teamId); err != nil {
					page.session.setFlashMessage("Error removing team: "+err.Error(), "error")
				} else {
					page.session.setFlashMessage("Team "+tm.Name+" Removed", "success")
				}
				redirect("/admin/teams", w, req)
			case "savemember":
				mbrName := req.FormValue("newmembername")
				mbr, err := NewTeamMember(tm.UUID, "")
				if err == nil {
					mbr.Name = mbrName
					mbr.SlackId = req.FormValue("newmemberslackid")
					mbr.Twitter = req.FormValue("newmembertwitter")
					mbr.Email = req.FormValue("newmemberemail")
				}
				if err := tm.AddTeamMember(mbr); err != nil {
					page.session.setFlashMessage("Error adding team member: "+err.Error(), "error")
				} else {
					page.session.setFlashMessage(mbrName+" added to team!", "success")
				}
				redirect("/admin/teams/"+teamId+"#members", w, req)
			case "deletemember":
				var err error
				var mbr *TeamMember
				if mbr, err = tm.GetTeamMemberById(req.FormValue("memberid")); err != nil {
					fmt.Println("Error removing team member: " + err.Error())
					page.session.setFlashMessage("Error deleting team member", "error")
					redirect("/admin/teams/"+teamId+"#members", w, req)
				}
				if err = tm.RemoveTeamMemberById(mbr.UUID); err != nil {
					fmt.Println("Error removing team member: " + err.Error())
					page.session.setFlashMessage("Error deleting team member", "error")
				} else {
					page.session.setFlashMessage(mbr.Name+" deleted from team", "success")
				}
				redirect("/admin/teams/"+teamId+"#members", w, req)
			default:
				page.SubTitle = "Edit Team"
				t, err := m.jam.GetTeamById(teamId)
				if err != nil {
					page.session.setFlashMessage("Error loading team: "+err.Error(), "error")
					redirect("/admin/teams", w, req)
				}
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
		page.TemplateData = m.jam
		page.SubTitle = "Teams"
		page.show("admin-teams.html", w)
	}
}
