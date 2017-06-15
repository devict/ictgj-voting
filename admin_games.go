package main

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

func handleAdminGames(w http.ResponseWriter, req *http.Request, page *pageData) {
	vars := mux.Vars(req)
	page.SubTitle = "Games"
	teamId := vars["id"]
	if teamId == "" {
		// Games List
		type gamesPageData struct {
			Games []Game
		}
		page.TemplateData = gamesPageData{Games: dbGetAllGames()}
		page.SubTitle = "Games"
		page.show("admin-games.html", w)
	} else {
		switch vars["function"] {
		case "save":
			name := req.FormValue("gamename")
			desc := req.FormValue("gamedesc")
			if dbIsValidTeam(teamId) {
				if err := dbUpdateTeamGame(teamId, name, desc); err != nil {
					page.session.setFlashMessage("Error updating game: "+err.Error(), "error")
				} else {
					page.session.setFlashMessage("Team game updated", "success")
				}
				redirect("/admin/teams/"+teamId, w, req)
			}
		case "screenshotupload":
			if err := saveScreenshots(teamId, req); err != nil {
				page.session.setFlashMessage("Error updating game: "+err.Error(), "error")
			}
			redirect("/admin/teams/"+teamId, w, req)
		}
	}
}

func saveScreenshots(teamId string, req *http.Request) error {
	err := req.ParseMultipartForm((1 << 10) * 24)
	if err != nil {
		return err
	}

	for _, fheaders := range req.MultipartForm.File {
		for _, hdr := range fheaders {
			// open uploaded
			var infile multipart.File
			if infile, err = hdr.Open(); err != nil {
				return err
			}

			// open destination
			var outfile *os.File
			if outfile, err = os.Create("./uploaded/" + hdr.Filename); err != nil {
				return err
			}
			// 32K buffer copy
			var written int64
			if written, err = io.Copy(outfile, infile); err != nil {
				return err
			}
			fmt.Println("uploaded file:" + hdr.Filename + ";length:" + strconv.Itoa(int(written)))
		}
	}
	return nil
}
