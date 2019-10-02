package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func handleAdminJam(w http.ResponseWriter, req *http.Request, page *pageData) {
	vars := mux.Vars(req)
	page.SubTitle = "Current Gamejam"
	fn := vars["id"]
	if fn == "save" {
		gjName := req.FormValue("jam_name")
		if gjName != "" {
			m.jam.Name = gjName
			err := m.saveChanges()
			if err == nil {
				page.session.setFlashMessage("Game Jam Updated", "success")
			} else {
				page.session.setFlashMessage("Error saving Game Jam", "error")
			}
		}
		redirect("/admin/jam", w, req)
	} else {
		page.TemplateData = m.jam
		page.show("admin-jam.html", w)
	}
}
