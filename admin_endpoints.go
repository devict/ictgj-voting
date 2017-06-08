package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func initAdminRequest(w http.ResponseWriter, req *http.Request) *pageData {
	p := InitPageData(w, req)
	p.Stylesheets = append(p.Stylesheets, "/assets/css/admin.css")
	p.Scripts = append(p.Scripts, "/assets/js/admin.js")
	p.HideAdminMenu = false

	return p
}

// Main admin handler, routes the request based on the category
func handleAdmin(w http.ResponseWriter, req *http.Request) {
	page := initAdminRequest(w, req)
	vars := mux.Vars(req)
	if !page.LoggedIn {
		if vars["category"] == "clients" &&
			vars["id"] != "" &&
			(vars["function"] == "add" || vars["function"] == "auth") {
			// When authenticating a client, we have an all-in-one login/auth page
			handleAdminClients(w, req, page)
		} else {
			page.SubTitle = "Admin Login"
			page.show("admin-login.html", w)
		}
	} else {
		adminCategory := vars["category"]
		switch adminCategory {
		case "users":
			handleAdminUsers(w, req, page)
		case "teams":
			handleAdminTeams(w, req, page)
		case "games":
			handleAdminGames(w, req, page)
		case "clients":
			handleAdminClients(w, req, page)
		default:
			page.show("admin-main.html", w)
		}
	}
}
