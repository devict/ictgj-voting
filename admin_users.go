package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// handleAdminDoLogin
// Verify the provided credentials, set up a cookie (if requested)
// and redirect back to /admin
// TODO: Set up the cookie
func handleAdminDoLogin(w http.ResponseWriter, req *http.Request) {
	page := initAdminRequest(w, req)
	// Fetch the login credentials
	email := req.FormValue("email")
	password := req.FormValue("password")
	if err := doLogin(email, password); err != nil {
		page.session.setFlashMessage("Invalid Login", "error")
	} else {
		page.session.setStringValue("email", email)
	}
	redirect("/admin", w, req)
}

// doLogin attempts to log in with the given email/password
// If it can't, it returns an error
func doLogin(email, password string) error {
	if strings.TrimSpace(email) != "" && strings.TrimSpace(password) != "" {
		return dbCheckCredentials(email, password)
	}
	return errors.New("Invalid Credentials")
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
