package main

import (
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
			page.SubTitle = "Authenticate Client"
			cli := dbGetClient(clientId)
			if cli.IP == "" {
				cli.IP = clientIp
			}
			type actClientPageData struct {
				Id   string
				Ip   string
				Name string
			}
			page.TemplateData = actClientPageData{Id: cli.UUID, Ip: cli.IP, Name: cli.Name}
			page.show("admin-activateclient.html", w)
		case "auth":
			email := req.FormValue("email")
			password := req.FormValue("password")
			clientName := req.FormValue("clientname")
			if clientName != "" {
				dbSetClientName(clientId, clientName)
			}
			dbUpdateClientIP(clientId, clientIp)
			if page.LoggedIn || doLogin(email, password) == nil {
				// Received a valid login
				// Authenticate the client
				if dbAuthClient(clientId, clientIp) == nil {
					page.session.setFlashMessage("Client Authenticated", "success")
				} else {
					page.session.setFlashMessage("Client Authentication Failed", "error")
				}
				if page.LoggedIn {
					redirect("/admin/clients", w, req)
				}
			}
			redirect("/", w, req)
		case "deauth":
			dbDeAuthClient(clientId)
			redirect("/admin/clients", w, req)
		}
	}
}

func clientIsAuthenticated(cid string, req *http.Request) bool {
	return dbClientIsAuth(cid)
	//return clientIsServer(req) || dbClientIsAuth(cid)
}

func clientIsServer(req *http.Request) bool {
	clientIp, _, _ := net.SplitHostPort(req.RemoteAddr)
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, i := range ifaces {
			if addrs, err := i.Addrs(); err == nil {
				for _, addr := range addrs {
					var ip net.IP
					switch v := addr.(type) {
					case *net.IPNet:
						ip = v.IP
					case *net.IPAddr:
						ip = v.IP
					}
					if clientIp == ip.String() {
						return true
					}
				}
			}
		}
	}
	return false
}
