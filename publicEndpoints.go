package main

import (
	"fmt"
	"net/http"
)

func initPublicPage(w http.ResponseWriter, req *http.Request) *pageData {
	p := InitPageData(w, req)
	return p
}

func handleMain(w http.ResponseWriter, req *http.Request) {
	page := initPublicPage(w, req)
	page.SubTitle = "!"
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
