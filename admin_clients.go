package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

func handleAdminClients(w http.ResponseWriter, req *http.Request, page *pageData) {
	vars := mux.Vars(req)
	page.SubTitle = "Clients"
	clientId := vars["id"]
	clientIp, _, _ := net.SplitHostPort(req.RemoteAddr)
	if clientId == "" {
		type clientsPageData struct {
			Clients []Client
		}
		page.TemplateData = clientsPageData{Clients: dbGetAllClients()}
		page.SubTitle = "Clients"
		page.show("admin-clients.html", w)
	} else {
		switch vars["function"] {
		case "add":
			type actClientPageData struct {
				Id string
			}
			page.TemplateData = actClientPageData{Id: clientId}
			page.show("admin-activateclient.html", w)
		case "auth":
			email := req.FormValue("email")
			password := req.FormValue("password")
			remember := req.FormValue("remember")
			if doLogin(email, password) == nil {
				// Received a valid login
				// Authenticate the client
				if dbAuthClient(clientId, clientIp) == nil {
					page.session.setFlashMessage("Client Authenticated", "success")
				} else {
					page.session.setFlashMessage("Client Authentication Failed", "error")
				}
				if remember == "on" {
					// Go ahead and log in
					page.session.setStringValue("email", email)
					redirect("/admin/clients", w, req)
				}
			}
			redirect("/", w, req)
		case "deauth":
			remember := req.FormValue("remember")
			fmt.Println("Remember: ", remember)
			dbDeAuthClient(clientId)
			redirect("/admin/clients", w, req)
		}
	}
}
