package main

import (
	"net/http"
)

func initPublicPage(w http.ResponseWriter, req *http.Request) *pageData {
	p := InitPageData(w, req)
	return p
}

func handleMain(w http.ResponseWriter, req *http.Request) {
	page := initPublicPage(w, req)
	page.SubTitle = ""
	switch dbGetPublicSiteMode() {
	case SiteModeWaiting:
		page.show("public-waiting.html", w)
	case SiteModeVoting:
		page.show("public-voting.html", w)
	}
}
