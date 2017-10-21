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
	client, err := m.GetClient(clientId)
	if err != nil {
		client = NewClient(clientId)
	}
	clientIp, _, _ := net.SplitHostPort(req.RemoteAddr)
	if clientId == "" {
		type clientsPageData struct {
			Clients []Client
		}
		page.TemplateData = clientsPageData{Clients: m.clients}
		page.SubTitle = "Clients"
		page.show("admin-clients.html", w)
	} else {
		switch vars["function"] {
		case "add":
			page.SubTitle = "Authenticate Client"
			if client.IP == "" {
				client.IP = clientIp
			}
			type actClientPageData struct {
				Id   string
				Ip   string
				Name string
			}
			page.TemplateData = actClientPageData{Id: client.UUID, Ip: client.IP, Name: client.Name}
			page.show("admin-activateclient.html", w)
		case "auth":
			email := req.FormValue("email")
			password := req.FormValue("password")
			clientName := req.FormValue("clientname")
			// Authentication isn't required to set a client name
			if clientName != "" {
				client.Name = clientName
			}
			client.IP = clientIp
			m.UpdateClient(client)
			if page.LoggedIn || doLogin(email, password) == nil {
				// Received a valid login
				// Authenticate the client
				client.Auth = true
				m.UpdateClient(client)
				page.session.setFlashMessage("Client Authenticated", "success")
				if page.LoggedIn {
					redirect("/admin/clients", w, req)
				}
			}
			redirect("/", w, req)
		case "deauth":
			client.Auth = false
			m.UpdateClient(client)
			page.session.setFlashMessage("Client De-Authenticated", "success")
			redirect("/admin/clients", w, req)
		}
	}
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
