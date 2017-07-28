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
	if dbGetPublicSiteMode() == SiteModeWaiting {
		page.SubTitle = ""
		page.show("public-waiting.html", w)
	} else {
		loadVotingPage(w, req)
	}
}

func loadVotingPage(w http.ResponseWriter, req *http.Request) {
	page := initPublicPage(w, req)
	// Client authentication required
	if (dbGetAuthMode() == AuthModeAuthentication) && !page.ClientIsAuth {
		page.show("unauthorized.html", w)
		return
	}
	type votingPageData struct {
		Teams     []Team
		Timestamp string
	}
	vpd := new(votingPageData)
	tms := dbGetAllTeams()

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
	if (dbGetAuthMode() == AuthModeAuthentication) && !page.ClientIsAuth {
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
	if _, err := dbGetVote(page.ClientId, timestamp); err == nil {
		// Duplicate vote... Cancel it.
		page.session.setFlashMessage("Duplicate vote!", "error")
		redirect("/", w, req)
	}
	// voteSlice is an ordered string slice of the voters preferences
	voteCSV := req.FormValue("uservote")
	voteSlice := strings.Split(voteCSV, ",")
	if err := dbSaveVote(page.ClientId, timestamp, voteSlice); err != nil {
		page.session.setFlashMessage("Error Saving Vote: "+err.Error(), "error")
	}
	if newVote, err := dbGetVote(page.ClientId, timestamp); err == nil {
		site.Votes = append(site.Votes, *newVote)
	}
	page.session.setFlashMessage("Vote Saved!", "success large fading")
	redirect("/", w, req)
}

func handleThumbnailRequest(w http.ResponseWriter, req *http.Request) {
	// Thumbnail requests are open even without client authentication
	vars := mux.Vars(req)
	ss := dbGetTeamGameScreenshot(vars["teamid"], vars["imageid"])
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
	ss := dbGetTeamGameScreenshot(vars["teamid"], vars["imageid"])
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
	if dbGetPublicSiteMode() == SiteModeVoting {
		redirect("/", w, req)
	}
	page := initPublicPage(w, req)
	vars := mux.Vars(req)
	page.SubTitle = "Team Details"
	teamId := vars["id"]
	if teamId != "" {
		// Team self-management functions
		if !dbIsValidTeam(teamId) {
			http.Error(w, "Page Not Found", 404)
			return
		}
		switch vars["function"] {
		case "":
			page.SubTitle = "Team Management"
			t := dbGetTeam(teamId)
			page.TemplateData = t
			page.show("public-teammgmt.html", w)
		case "savemember":
			mbrName := req.FormValue("newmembername")
			mbrSlack := req.FormValue("newmemberslackid")
			mbrTwitter := req.FormValue("newmembertwitter")
			mbrEmail := req.FormValue("newmemberemail")
			if err := dbAddTeamMember(teamId, mbrName, mbrEmail, mbrSlack, mbrTwitter); err != nil {
				page.session.setFlashMessage("Error adding team member: "+err.Error(), "error")
			} else {
				page.session.setFlashMessage(mbrName+" added to team!", "success")
			}
			refreshTeamsInMemory()
			redirect("/team/"+teamId+"#members", w, req)
		case "deletemember":
			mbrId := req.FormValue("memberid")
			m, _ := dbGetTeamMember(teamId, mbrId)
			if err := dbDeleteTeamMember(teamId, mbrId); err != nil {
				page.session.setFlashMessage("Error deleting team member: "+err.Error(), "error")
			} else {
				page.session.setFlashMessage(m.Name+" deleted from team", "success")
			}
			refreshTeamsInMemory()
			redirect("/team/"+teamId, w, req)
		case "savegame":
			name := req.FormValue("gamename")
			link := req.FormValue("gamelink")
			desc := req.FormValue("gamedesc")
			if dbIsValidTeam(teamId) {
				if err := dbUpdateTeamGame(teamId, name, link, desc); err != nil {
					page.session.setFlashMessage("Error updating game: "+err.Error(), "error")
				} else {
					page.session.setFlashMessage("Team game updated", "success")
				}
				redirect("/team/"+teamId, w, req)
			}
		case "screenshotupload":
			if err := saveScreenshots(teamId, req); err != nil {
				page.session.setFlashMessage("Error updating game: "+err.Error(), "error")
			}
			redirect("/team/"+teamId, w, req)
		case "screenshotdelete":
			ssid := vars["subid"]
			if err := dbDeleteTeamGameScreenshot(teamId, ssid); err != nil {
				page.session.setFlashMessage("Error deleting screenshot: "+err.Error(), "error")
			}
			redirect("/team/"+teamId, w, req)
		}
	}
}
