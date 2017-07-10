package main

import (
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func initPublicPage(w http.ResponseWriter, req *http.Request) *pageData {
	p := InitPageData(w, req)
	return p
}

func handleMain(w http.ResponseWriter, req *http.Request) {
	switch dbGetPublicSiteMode() {
	case SiteModeWaiting:
		page := initPublicPage(w, req)
		page.SubTitle = ""
		page.show("public-waiting.html", w)
	case SiteModeVoting:
		loadVotingPage(w, req)
	}
}

func loadVotingPage(w http.ResponseWriter, req *http.Request) {
	page := initPublicPage(w, req)
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
