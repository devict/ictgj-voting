package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func initAdminRequest(w http.ResponseWriter, req *http.Request) *pageData {
	if site.DevMode {
		w.Header().Set("Cache-Control", "no-cache")
	}
	p := new(pageData)
	// Get session
	var err error
	var s *sessions.Session
	if s, err = sessionStore.Get(req, site.SessionName); err != nil {
		http.Error(w, err.Error(), 500)
		return p
	}
	p.session = new(pageSession)
	p.session.session = s
	p.session.req = req
	p.session.w = w

	// First check if we're logged in
	userEmail, _ := p.session.getStringValue("email")

	// With a valid account
	p.LoggedIn = dbIsValidUserEmail(userEmail)
	p.Site = site
	p.SubTitle = ""
	p.Stylesheets = make([]string, 0, 0)
	p.Stylesheets = append(p.Stylesheets, "/assets/css/pure-min.css")
	p.Stylesheets = append(p.Stylesheets, "/assets/css/grids-responsive-min.css")
	p.Stylesheets = append(p.Stylesheets, "/assets/font-awesome/css/font-awesome.min.css")
	p.Stylesheets = append(p.Stylesheets, "/assets/css/gjvote.css")
	p.Stylesheets = append(p.Stylesheets, "/assets/css/admin.css")

	p.HeaderScripts = make([]string, 0, 0)
	p.HeaderScripts = append(p.HeaderScripts, "/assets/js/snack-min.js")
	p.Scripts = make([]string, 0, 0)
	p.Scripts = append(p.Scripts, "/assets/js/admin.js")
	p.FlashMessage, p.FlashClass = p.session.getFlashMessage()
	if p.FlashClass == "" {
		p.FlashClass = "hidden"
	}
	// Build the menu
	if p.LoggedIn {
		p.Menu = append(p.Menu, menuItem{"Votes", "/admin/votes", "fa-sticky-note"})
		p.Menu = append(p.Menu, menuItem{"Teams", "/admin/teams", "fa-users"})
		p.Menu = append(p.Menu, menuItem{"Games", "/admin/games", "fa-gamepad"})

		p.BottomMenu = append(p.BottomMenu, menuItem{"Users", "/admin/users", "fa-user"})
		p.BottomMenu = append(p.BottomMenu, menuItem{"Logout", "/admin/dologout", "fa-sign-out"})
	}
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
}

// handleAdminGames
func handleAdminGames(w http.ResponseWriter, req *http.Request, page *pageData) {
}
