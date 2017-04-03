package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

func initPublicPage(w http.ResponseWriter, req *http.Request) *pageData {
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

	p.Site = site
	p.SubTitle = "GameJam Voting"
	p.Stylesheets = make([]string, 0, 0)
	p.Stylesheets = append(p.Stylesheets, "/assets/css/pure-min.css")
	p.Stylesheets = append(p.Stylesheets, "/assets/font-awesome/css/font-awesome.min.css")
	p.Stylesheets = append(p.Stylesheets, "/assets/css/gjvote.css")
	p.HeaderScripts = make([]string, 0, 0)
	p.HeaderScripts = append(p.HeaderScripts, "/assets/js/snack-min.js")
	p.Scripts = make([]string, 0, 0)
	p.Scripts = append(p.Scripts, "/assets/js/gjvote.js")
	p.FlashMessage, p.FlashClass = p.session.getFlashMessage()
	return p
}

func handleMain(w http.ResponseWriter, req *http.Request) {
	page := initPublicPage(w, req)
	page.SubTitle = "Place your Vote!"
	for _, tmpl := range []string{
		"htmlheader.html",
		"main.html",
		"footer.html",
		"htmlfooter.html",
	} {
		if err := outputTemplate(tmpl, page, w); err != nil {
			fmt.Printf("%s\n", err)
		}
	}
}
