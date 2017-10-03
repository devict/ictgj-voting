package main

import (
	"encoding/base64"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func initPublicPage(w http.ResponseWriter, req *http.Request) *pageData {
	p := InitPageData(w, req)
	return p
}

func handleMain(w http.ResponseWriter, req *http.Request) {
	page := initPublicPage(w, req)
	if db.getPublicSiteMode() == SiteModeWaiting {
		page.SubTitle = ""
		page.show("public-waiting.html", w)
	} else {
		loadVotingPage(w, req)
	}
}

func loadVotingPage(w http.ResponseWriter, req *http.Request) {
	page := initPublicPage(w, req)
	// Client authentication required
	if (db.getAuthMode() == AuthModeAuthentication) && !page.ClientIsAuth {
		page.show("unauthorized.html", w)
		return
	}
	type votingPageData struct {
		Teams     []Team
		Timestamp string
	}
	vpd := new(votingPageData)
	tms := db.getAllTeams()

	// Randomize the team list
	rand.Seed(time.Now().Unix())
	for len(tms) > 0 {
		i := rand.Intn(len(tms))
		vpd.Teams = append(vpd.Teams, tms[i])
		tms = append(tms[:i], tms[i+1:]...)
	}

	vpd.Timestamp = time.Now().Format(time.RFC3339)
	page.TemplateData = vpd
	page.show("public-voting.html", w)
}

func handlePublicSaveVote(w http.ResponseWriter, req *http.Request) {
	page := initPublicPage(w, req)
	// Client authentication required
	if (db.getAuthMode() == AuthModeAuthentication) && !page.ClientIsAuth {
		page.show("unauthorized.html", w)
		return
	}

	page.SubTitle = ""

	// Check if we already have a vote for this client id/timestamp
	ts := req.FormValue("timestamp")
	timestamp, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		page.session.setFlashMessage("Error parsing timestamp: "+ts, "error")
		redirect("/", w, req)
	}
	client := db.getClient(page.ClientId)
	if _, err := client.getVote(timestamp); err == nil {
		// Duplicate vote... Cancel it.
		page.session.setFlashMessage("Duplicate vote!", "error")
		redirect("/", w, req)
	}
	// voteSlice is an ordered string slice of the voters preferences
	voteCSV := req.FormValue("uservote")
	voteSlice := strings.Split(voteCSV, ",")
	if err := client.saveVote(timestamp, voteSlice); err != nil {
		page.session.setFlashMessage("Error Saving Vote: "+err.Error(), "error")
	}
	if newVote, err := client.getVote(timestamp); err == nil {
		site.Votes = append(site.Votes, *newVote)
	}
	page.session.setFlashMessage("Vote Saved!", "success large fading")
	redirect("/", w, req)
}

func handleThumbnailRequest(w http.ResponseWriter, req *http.Request) {
	// Thumbnail requests are open even without client authentication
	vars := mux.Vars(req)
	tm := db.getTeam(vars["teamid"])
	if tm == nil {
		http.Error(w, "Couldn't find image", 404)
		return
	}
	ss := tm.getScreenshot(vars["imageid"])
	if ss == nil {
		http.Error(w, "Couldn't find image", 404)
		return
	}
	w.Header().Set("Content-Type", "image/"+ss.Filetype)
	dat, err := base64.StdEncoding.DecodeString(ss.Thumbnail)
	if err != nil {
		http.Error(w, "Couldn't find image", 404)
		return
	}
	w.Write(dat)
}

func handleImageRequest(w http.ResponseWriter, req *http.Request) {
	// Image requests are open even without client authentication
	vars := mux.Vars(req)
	tm := db.getTeam(vars["teamid"])
	if tm == nil {
		http.Error(w, "Couldn't find image", 404)
		return
	}
	ss := tm.getScreenshot(vars["imageid"])
	if ss == nil {
		http.Error(w, "Couldn't find image", 404)
		return
	}
	w.Header().Set("Content-Type", "image/"+ss.Filetype)
	dat, err := base64.StdEncoding.DecodeString(ss.Image)
	if err != nil {
		http.Error(w, "Couldn't find image", 404)
		return
	}
	w.Write(dat)
}

func handleTeamMgmtRequest(w http.ResponseWriter, req *http.Request) {
	// Team Management pages are open even without client authentication
	if db.getPublicSiteMode() == SiteModeVoting {
		redirect("/", w, req)
	}
	page := initPublicPage(w, req)
	vars := mux.Vars(req)
	page.SubTitle = "Team Details"
	teamId := vars["id"]
	tm := db.getTeam(teamId)
	if tm != nil {
		// Team self-management functions
		switch vars["function"] {
		case "":
			page.SubTitle = "Team Management"
			page.TemplateData = tm
			page.show("public-teammgmt.html", w)
		case "savemember":
			m := newTeamMember(req.FormValue("newmembername"))
			m.SlackId = req.FormValue("newmemberslackid")
			m.Twitter = req.FormValue("newmembertwitter")
			m.Email = req.FormValue("newmemberemail")
			if err := tm.updateTeamMember(m); err != nil {
				page.session.setFlashMessage("Error adding team member: "+err.Error(), "error")
			} else {
				page.session.setFlashMessage(m.Name+" added to team!", "success")
			}
			refreshTeamsInMemory()
			redirect("/team/"+tm.UUID+"#members", w, req)
		case "deletemember":
			mbrId := req.FormValue("memberid")
			m := tm.getTeamMember(mbrId)
			if m != nil {
				if err := tm.deleteTeamMember(m); err != nil {
					page.session.setFlashMessage("Error deleting team member: "+err.Error(), "error")
				} else {
					page.session.setFlashMessage(m.Name+" deleted from team", "success")
				}
			} else {
				page.session.setFlashMessage("Couldn't find member to delete", "error")
			}
			refreshTeamsInMemory()
			redirect("/team/"+tm.UUID, w, req)
		case "savegame":
			gm := newGame(tm.UUID)
			gm.Name = req.FormValue("gamename")
			gm.Link = req.FormValue("gamelink")
			gm.Description = req.FormValue("gamedesc")
			if err := gm.save(); err != nil {
				page.session.setFlashMessage("Error updating game: "+err.Error(), "error")
			} else {
				page.session.setFlashMessage("Team game updated", "success")
			}
			redirect("/team/"+tm.UUID, w, req)
		case "screenshotupload":
			if err := saveScreenshots(tm, req); err != nil {
				page.session.setFlashMessage("Error updating game: "+err.Error(), "error")
			}
			redirect("/team/"+tm.UUID, w, req)
		case "screenshotdelete":
			ssid := vars["subid"]
			if err := tm.deleteScreenshot(ssid); err != nil {
				page.session.setFlashMessage("Error deleting screenshot: "+err.Error(), "error")
			}
			redirect("/team/"+tm.UUID, w, req)
		}
	} else {
		http.Error(w, "Page Not Found", 404)
		return
	}
}
