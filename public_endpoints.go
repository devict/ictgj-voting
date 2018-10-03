package main

import (
	"encoding/base64"
	"fmt"
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
	if m.site.GetPublicMode() == SiteModeWaiting {
		page.SubTitle = ""
		page.show("public-waiting.html", w)
	} else {
		loadVotingPage(w, req)
	}
}

func loadVotingPage(w http.ResponseWriter, req *http.Request) {
	page := initPublicPage(w, req)
	// Client authentication required
	if (m.site.GetAuthMode() == AuthModeAuthentication) && !page.ClientIsAuth {
		page.show("unauthorized.html", w)
		return
	}
	type votingPageData struct {
		Teams     []Team
		Timestamp string
	}
	vpd := new(votingPageData)
	tms := make([]Team, len(m.jam.Teams))
	copy(tms, m.jam.Teams)

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
	if (m.site.GetAuthMode() == AuthModeAuthentication) && !page.ClientIsAuth {
		page.show("unauthorized.html", w)
		return
	}

	page.SubTitle = ""

	// Check if we already have a vote for this client id/timestamp
	ts := req.FormValue("timestamp")
	timestamp, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		page.session.setFlashMessage("Error creating vote", "error")
		fmt.Println("Error parsing timestamp: " + ts)
		redirect("/", w, req)
	}
	client, err := m.GetClient(page.ClientId)
	if err != nil {
		client = NewClient(page.ClientId)
	}

	// voteSlice is an ordered string slice of the voters preferences
	voteCSV := req.FormValue("uservote")
	voteSlice := strings.Split(voteCSV, ",")

	// Voter Status should be either 'participant', 'volunteer', or 'visitor'
	voterStatus := req.FormValue("voterstatus")
	// Discovery is how the voter found out about the gamejam
	discovery := req.FormValue("discovery")

	if _, err = m.jam.GetVote(client.UUID, timestamp); err == nil {
		// Duplicate vote... Cancel it.
		page.session.setFlashMessage("Duplicate vote!", "error")
		redirect("/", w, req)
	}

	var vt *Vote
	if vt, err = NewVote(client.UUID, timestamp); err != nil {
		fmt.Println("Error creating vote: " + err.Error())
		page.session.setFlashMessage("Error creating vote", "error")
		redirect("/", w, req)
	}
	if err = vt.SetChoices(voteSlice); err != nil {
		fmt.Println("Error creating vote: " + err.Error())
		page.session.setFlashMessage("Error creating vote", "error")
		redirect("/", w, req)
	}
	vt.VoterStatus = voterStatus
	vt.Discovery = discovery

	if err := m.jam.AddVote(vt); err != nil {
		fmt.Println("Error adding vote: " + err.Error())
		page.session.setFlashMessage("Error creating vote", "error")
		redirect("/", w, req)
	}
	page.session.setFlashMessage("Vote Saved!", "success large fading")
	redirect("/", w, req)
}

func handleThumbnailRequest(w http.ResponseWriter, req *http.Request) {
	// Thumbnail requests are open even without client authentication
	vars := mux.Vars(req)
	tm, err := m.jam.GetTeamById(vars["teamid"])
	if err != nil {
		fmt.Println("handleThumbnailRequest: " + err.Error())
		http.Error(w, "Couldn't find image", 404)
		return
	}
	ss, err := tm.Game.GetScreenshot(vars["imageid"])
	if err != nil {
		fmt.Println("handleThumbnailRequest: " + err.Error())
		http.Error(w, "Couldn't find image", 404)
		return
	}
	w.Header().Set("Content-Type", "image/"+ss.Filetype)
	dat, err := base64.StdEncoding.DecodeString(ss.Thumbnail)
	if err != nil {
		fmt.Println("handleThumbnailRequest: " + err.Error())
		http.Error(w, "Couldn't find image", 404)
		return
	}
	w.Write(dat)
}

func handleImageRequest(w http.ResponseWriter, req *http.Request) {
	// Image requests are open even without client authentication
	vars := mux.Vars(req)
	tm, err := m.jam.GetTeamById(vars["teamid"])
	if err != nil {
		fmt.Println("handleImageRequest: " + err.Error())
		http.Error(w, "Couldn't find image", 404)
		return
	}
	ss, err := tm.Game.GetScreenshot(vars["imageid"])
	if err != nil {
		fmt.Println("handleImageRequest: " + err.Error())
		http.Error(w, "Couldn't find image", 404)
		return
	}
	w.Header().Set("Content-Type", "image/"+ss.Filetype)
	dat, err := base64.StdEncoding.DecodeString(ss.Image)
	if err != nil {
		fmt.Println("handleImageRequest: " + err.Error())
		http.Error(w, "Couldn't find image", 404)
		return
	}
	w.Write(dat)
}

func handleTeamMgmtRequest(w http.ResponseWriter, req *http.Request) {
	// Team Management pages are open even without client authentication
	if m.site.GetPublicMode() == SiteModeVoting {
		redirect("/", w, req)
	}
	page := initPublicPage(w, req)
	vars := mux.Vars(req)
	page.SubTitle = "Team Details"
	teamId := vars["id"]
	tm, err := m.jam.GetTeamById(teamId)
	if err == nil {
		// Team self-management functions
		switch vars["function"] {
		case "":
			page.SubTitle = "Team Management"
			page.TemplateData = tm
			page.show("public-teammgmt.html", w)

		case "savemember":
			m, err := NewTeamMember(tm.UUID, "")
			if err != nil {
				page.session.setFlashMessage("Error adding team member: "+err.Error(), "error")
				redirect("/team/"+tm.UUID+"#members", w, req)
			}
			m.Name = req.FormValue("newmembername")
			m.SlackId = req.FormValue("newmemberslackid")
			m.Twitter = req.FormValue("newmembertwitter")
			m.Email = req.FormValue("newmemberemail")
			if err := tm.AddTeamMember(m); err != nil {
				page.session.setFlashMessage("Error adding team member: "+err.Error(), "error")
			} else {
				page.session.setFlashMessage(m.Name+" added to team!", "success")
			}
			redirect("/team/"+tm.UUID+"#members", w, req)

		case "deletemember":
			mbrId := req.FormValue("memberid")
			err := tm.RemoveTeamMemberById(mbrId)
			if err != nil {
				page.session.setFlashMessage("Error deleting team member: "+err.Error(), "error")
			} else {
				page.session.setFlashMessage("Team member removed", "success")
			}
			redirect("/team/"+tm.UUID, w, req)

		case "savegame":
			tm.Game.Name = req.FormValue("gamename")
			tm.Game.Link = req.FormValue("gamelink")
			tm.Game.Description = req.FormValue("gamedesc")
			page.session.setFlashMessage("Team game updated", "success")
			redirect("/team/"+tm.UUID, w, req)

		case "screenshotupload":
			ss, err := ssFromRequest(tm, req)
			if err != nil {
				page.session.setFlashMessage("Error updating game: "+err.Error(), "error")
				redirect("/team/"+tm.UUID, w, req)
			}
			gm := tm.Game
			gm.Screenshots = append(gm.Screenshots, *ss)
			if err = m.jam.UpdateGame(tm.UUID, gm); err != nil {
				page.session.setFlashMessage("Error updating game: "+err.Error(), "error")
			} else {
				page.session.setFlashMessage("Screenshot Uploaded", "success")
			}
			redirect("/team/"+tm.UUID, w, req)

		case "screenshotdelete":
			ssid := vars["subid"]
			if err := tm.Game.RemoveScreenshot(ssid); err != nil {
				page.session.setFlashMessage("Error deleting screenshot: "+err.Error(), "error")
			}
			redirect("/team/"+tm.UUID, w, req)

		}
	} else {
		http.Error(w, "Page Not Found", 404)
		return
	}
}
