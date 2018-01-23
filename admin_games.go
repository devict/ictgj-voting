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
		page.TemplateData = m.jam
		page.SubTitle = "Games"
		page.show("admin-games.html", w)
	} else {
		tm, _ := m.jam.GetTeamById(teamId)
		if tm != nil {
			switch vars["function"] {
			case "save":
				var err error
				var gm *Game
				if gm, err = NewGame(tm.UUID); err != nil {
					page.session.setFlashMessage("Error updating game: "+err.Error(), "error")
				}
				gm.Name = req.FormValue("gamename")
				gm.Link = req.FormValue("gamelink")
				gm.Description = req.FormValue("gamedesc")
				if err := m.jam.UpdateGame(tm.UUID, gm); err != nil {
					page.session.setFlashMessage("Error updating game: "+err.Error(), "error")
				} else {
					page.session.setFlashMessage("Team game updated", "success")
				}
				redirect("/admin/teams/"+tm.UUID+"#game", w, req)

			case "screenshotupload":
				var ss *Screenshot
				tm, err := m.jam.GetTeamById(tm.UUID)
				if err != nil {
					page.session.setFlashMessage("Error updating game: "+err.Error(), "error")
					redirect("/admin/teams/"+tm.UUID+"#game", w, req)
				}
				ss, err = ssFromRequest(tm, req)
				if err != nil {
					page.session.setFlashMessage("Error updating game: "+err.Error(), "error")
					redirect("/admin/teams/"+tm.UUID+"#game", w, req)
				}
				gm := tm.Game
				gm.Screenshots = append(gm.Screenshots, *ss)
				if err = m.jam.UpdateGame(tm.UUID, gm); err != nil {
					page.session.setFlashMessage("Error updating game: "+err.Error(), "error")
				} else {
					page.session.setFlashMessage("Screenshot Uploaded", "success")
				}
				redirect("/admin/teams/"+tm.UUID+"#game", w, req)

			case "screenshotdelete":
				var err error
				ssid := vars["subid"]
				tm, err := m.jam.GetTeamById(tm.UUID)
				if err != nil {
					page.session.setFlashMessage("Error updating game: "+err.Error(), "error")
					redirect("/admin/teams/"+tm.UUID+"#game", w, req)
					break
				}
				gm := tm.Game
				if err = gm.RemoveScreenshot(ssid); err != nil {
					page.session.setFlashMessage("Error updating game: "+err.Error(), "error")
					redirect("/admin/teams/"+tm.UUID+"#game", w, req)
					break
				}
				if err = m.jam.UpdateGame(tm.UUID, gm); err != nil {
					page.session.setFlashMessage("Error updating game: "+err.Error(), "error")
				} else {
					page.session.setFlashMessage("Screenshot Removed", "success")
				}
				redirect("/admin/teams/"+tm.UUID+"#game", w, req)

			}
		} else {
			page.session.setFlashMessage("Not a valid team id", "error")
			redirect("/admin/teams", w, req)
		}
	}
}

func ssFromRequest(tm *Team, req *http.Request) (*Screenshot, error) {
	var err error
	var ss *Screenshot

	file, hdr, err := req.FormFile("newssfile")
	if err != nil {
		return nil, err
	}
	extIdx := strings.LastIndex(hdr.Filename, ".")
	fltp := "png"
	if len(hdr.Filename) > extIdx {
		fltp = hdr.Filename[extIdx+1:]
	}
	mI, _, err := image.Decode(file)
	buf := new(bytes.Buffer)
	// We convert everything to jpg
	if err = jpeg.Encode(buf, mI, nil); err != nil {
		return nil, errors.New("Unable to encode image")
	}
	thm := resize.Resize(200, 0, mI, resize.Lanczos3)
	thmBuf := new(bytes.Buffer)
	var thmString string
	if fltp == "gif" {
		if err = gif.Encode(thmBuf, thm, nil); err != nil {
			return nil, errors.New("Unable to encode image")
		}
	} else {
		if err = jpeg.Encode(thmBuf, thm, nil); err != nil {
			return nil, errors.New("Unable to encode image")
		}
	}
	thmString = base64.StdEncoding.EncodeToString(thmBuf.Bytes())

	if ss, err = NewScreenshot(tm.UUID, ""); err != nil {
		return nil, err
	}

	ss.Image = base64.StdEncoding.EncodeToString(buf.Bytes())
	ss.Thumbnail = thmString
	ss.Filetype = fltp

	return ss, nil
	//return m.jam.SaveScreenshot(ss)
}
