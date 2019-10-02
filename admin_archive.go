package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func handleAdminArchive(w http.ResponseWriter, req *http.Request, page *pageData) {
	vars := mux.Vars(req)
	page.SubTitle = "GameJam Archive"
	id := vars["id"]
	if id == "archive-current" {
		// Archive the current gamejam
		if err := m.ArchiveCurrentJam(); err != nil {
			page.session.setFlashMessage("Error archiving jam", "error")
			fmt.Println(err.Error())
		}
		redirect("/admin/jam", w, req)
	} else if id != "" {
		// Display a specific archive
		agj := new(ArchivedGamejam)
		for _, v := range m.archive.Jams {
			if id == v.UUID {
				agj.UUID = v.UUID
				agj.Name = v.Name
				agj.Date = v.Date
				agj.Rankings = v.Rankings
				agj.Teams = v.Teams
				agj.Votes = v.Votes
				break
			}
		}
		// We want to replace the team UUIDs in the rankings with their name
		for k, v := range agj.Rankings {
			for _, tv := range agj.Teams {
				if tv.UUID == v {
					agj.Rankings[k] = tv.Name
					break
				}
			}
		}
		page.TemplateData = agj
		page.SubTitle = "Archived Game Jam"
		page.show("admin-viewarchived.html", w)
	} else {
		// Archive List
		page.TemplateData = m.archive
		page.SubTitle = "Archive"
		page.show("admin-archive.html", w)
	}
}
