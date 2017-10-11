package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
)

func handleAdminGames(w http.ResponseWriter, req *http.Request, page *pageData) {
	vars := mux.Vars(req)
	page.SubTitle = "Games"
	teamId := vars["id"]
	if teamId == "" {
		// Games List
		type gamesPageData struct {
			Teams []Team
		}
		gpd := new(gamesPageData)
		gpd.Teams = db.getAllTeams()
		page.TemplateData = gpd
		page.SubTitle = "Games"
		page.show("admin-games.html", w)
	} else {
		tm := db.getTeam(teamId)
		if tm != nil {
			switch vars["function"] {
			case "save":
				gm := db.newGame(tm.UUID)
				gm.Name = req.FormValue("gamename")
				gm.Link = req.FormValue("gamelink")
				gm.Description = req.FormValue("gamedesc")
				if err := gm.save(); err != nil {
					page.session.setFlashMessage("Error updating game: "+err.Error(), "error")
				} else {
					page.session.setFlashMessage("Team game updated", "success")
				}
				redirect("/admin/teams/"+tm.UUID+"#game", w, req)
			case "screenshotupload":
				if err := saveScreenshots(tm, req); err != nil {
					page.session.setFlashMessage("Error updating game: "+err.Error(), "error")
				}
				redirect("/admin/teams/"+tm.UUID+"#game", w, req)
			case "screenshotdelete":
				ssid := vars["subid"]
				if err := tm.deleteScreenshot(ssid); err != nil {
					page.session.setFlashMessage("Error deleting screenshot: "+err.Error(), "error")
				}
				redirect("/admin/teams/"+tm.UUID+"#game", w, req)
			}
		} else {
			page.session.setFlashMessage("Not a valid team id", "error")
			redirect("/admin/teams", w, req)
		}
	}
}

func saveScreenshots(tm *Team, req *http.Request) error {
	var err error
	file, hdr, err := req.FormFile("newssfile")
	if err != nil {
		return err
	}
	extIdx := strings.LastIndex(hdr.Filename, ".")
	fltp := "png"
	if len(hdr.Filename) > extIdx {
		fltp = hdr.Filename[extIdx+1:]
	}
	m, _, err := image.Decode(file)
	buf := new(bytes.Buffer)
	// We convert everything to jpg
	if err = jpeg.Encode(buf, m, nil); err != nil {
		return errors.New("Unable to encode image")
	}
	thm := resize.Resize(200, 0, m, resize.Lanczos3)
	thmBuf := new(bytes.Buffer)
	var thmString string
	if fltp == "gif" {
		if err = gif.Encode(thmBuf, thm, nil); err != nil {
			return errors.New("Unable to encode image")
		}
	} else {
		if err = jpeg.Encode(thmBuf, thm, nil); err != nil {
			return errors.New("Unable to encode image")
		}
	}
	thmString = base64.StdEncoding.EncodeToString(thmBuf.Bytes())

	return tm.saveScreenshot(&Screenshot{
		Image:     base64.StdEncoding.EncodeToString(buf.Bytes()),
		Thumbnail: thmString,
		Filetype:  fltp,
	})
}
