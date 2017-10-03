package main

import (
	"net/http"
	"strconv"

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
		case "votes":
			handleAdminVotes(w, req, page)
		case "mode":
			handleAdminSetMode(w, req, page)
		case "authmode":
			handleAdminSetAuthMode(w, req, page)
		case "archive":
			handleAdminArchive(w, req, page)
		default:
			page.TemplateData = getCondorcetResult()
			page.show("admin-main.html", w)
		}
	}
}

func handleAdminSetMode(w http.ResponseWriter, req *http.Request, page *pageData) {
	vars := mux.Vars(req)
	newMode, err := strconv.Atoi(vars["id"])
	if err != nil {
		page.session.setFlashMessage("Invalid Mode: "+vars["id"], "error")
	}
	if dbSetPublicSiteMode(newMode) != nil {
		page.session.setFlashMessage("Invalid Mode: "+vars["id"], "error")
	}
	redirect("/admin", w, req)
}

func handleAdminSetAuthMode(w http.ResponseWriter, req *http.Request, page *pageData) {
	vars := mux.Vars(req)
	newMode, err := strconv.Atoi(vars["id"])
	if err != nil {
		page.session.setFlashMessage("Invalid Authentication Mode: "+vars["id"], "error")
	}
	if dbSetAuthMode(newMode) != nil {
		page.session.setFlashMessage("Invalid Authentication Mode: "+vars["id"], "error")
	}
	redirect("/admin", w, req)
}
