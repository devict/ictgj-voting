package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func initAdminRequest(w http.ResponseWriter, req *http.Request) *pageData {
	p := InitPageData(w, req)
	p.Stylesheets = append(p.Stylesheets, "/assets/css/admin.css")
	p.Scripts = append(p.Scripts, "/assets/js/admin.js")

	return p
}

// handleAdmin
// Main admin handler, routes the request based on the category
func handleAdmin(w http.ResponseWriter, req *http.Request) {
	page := initAdminRequest(w, req)
	if !page.LoggedIn {
		page.SubTitle = "Admin Login"
		page.show("admin-login.html", w)
	} else {
		vars := mux.Vars(req)
		adminCategory := vars["category"]
		switch adminCategory {
		case "users":
			handleAdminUsers(w, req, page)
		case "teams":
			handleAdminTeams(w, req, page)
		case "games":
			handleAdminGames(w, req, page)
		default:
			page.show("admin-main.html", w)
		}
	}
}

// handleAdminDoLogin
// Verify the provided credentials, set up a cookie (if requested)
// and redirect back to /admin
func handleAdminDoLogin(w http.ResponseWriter, req *http.Request) {
	page := initAdminRequest(w, req)
	// Fetch the login credentials
	email := req.FormValue("email")
	password := req.FormValue("password")
	if email != "" && password != "" {
		if err := dbCheckCredentials(email, password); err != nil {
			page.session.setFlashMessage("Invalid Login", "error")
		} else {
			page.session.setStringValue("email", email)
		}
	} else {
		page.session.setFlashMessage("Invalid Login", "error")
	}
	redirect("/admin", w, req)
}

// handleAdminDoLogout
// Expire the session
func handleAdminDoLogout(w http.ResponseWriter, req *http.Request) {
	page := initAdminRequest(w, req)
	page.session.expireSession()
	page.session.setFlashMessage("Logged Out", "success")

	redirect("/admin", w, req)
}

// handleAdminUsers
func handleAdminUsers(w http.ResponseWriter, req *http.Request, page *pageData) {
	vars := mux.Vars(req)
	page.SubTitle = "Admin Users"
	email := vars["id"]
	if email == "new" {
		switch vars["function"] {
		case "save":
			email = req.FormValue("email")
			if dbIsValidUserEmail(email) {
				// User already exists
				page.session.setFlashMessage("A user with email address "+email+" already exists!", "error")
			} else {
				password := req.FormValue("password")
				if err := dbUpdateUserPassword(email, string(password)); err != nil {
					page.session.setFlashMessage(err.Error(), "error")
				} else {
					page.session.setFlashMessage("User "+email+" created!", "success")
				}
			}
			redirect("/admin/users", w, req)
		default:
			page.SubTitle = "Add Admin User"
			page.show("admin-adduser.html", w)
		}
	} else if email != "" {
		switch vars["function"] {
		case "save":
			var err error
			if dbIsValidUserEmail(email) {
				password := req.FormValue("password")
				if password != "" {
					if err = dbUpdateUserPassword(email, password); err != nil {
						page.session.setFlashMessage(err.Error(), "error")
					} else {
						page.session.setFlashMessage("User "+email+" created!", "success")
					}
				}
				redirect("/admin/users", w, req)
			}
		case "delete":
			var err error
			if dbIsValidUserEmail(email) {
				if err = dbDeleteUser(email); err != nil {
					page.session.setFlashMessage(err.Error(), "error")
				} else {
					page.session.setFlashMessage("User "+email+" deleted!", "success")
				}
			}
			redirect("/admin/users", w, req)
		default:
			page.SubTitle = "Edit Admin User"
			if !dbIsValidUserEmail(email) {
				page.session.setFlashMessage("Couldn't find the requested user, please try again.", "error")
				redirect("/admin/users", w, req)
			}
			page.TemplateData = email
			page.show("admin-edituser.html", w)
		}
	} else {
		type usersPageData struct {
			Users []string
		}
		page.TemplateData = usersPageData{Users: dbGetAllUsers()}

		page.SubTitle = "Admin Users"
		page.show("admin-users.html", w)
	}
}

// handleAdminTeams
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
				page.session.setFlashMessage("Not implemented yet...", "success")
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

// handleAdminGames
func handleAdminGames(w http.ResponseWriter, req *http.Request, page *pageData) {
}
