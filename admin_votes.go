package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func getCondorcetResult() []Team {
	var ret []Team
	type rankedTeam struct {
		tm     *Team
		wins   map[string]int
		losses map[string]int
	}
	// Build our Ranked Teams slice
	allRanks := make(map[string]rankedTeam)
	for i := range site.Teams {
		rt := new(rankedTeam)
		rt.wins = make(map[string]int)
		rt.losses = make(map[string]int)
		rt.tm = &site.Teams[i]
		for j := range site.Teams {
			if site.Teams[i].UUID != site.Teams[j].UUID {
				rt.wins[site.Teams[j].UUID] = 0
				rt.losses[site.Teams[j].UUID] = 0
			}
		}
		allRanks[site.Teams[i].UUID] = *rt
	}
	/*
		Go through all votes, for each choice (ct):
		* Go through all teams (tm)
			* if tm was processed earlier, do nothing
			* otherwise mark a win for ct against tm, and a loss for tm against cm
	*/
	// Now go through all of the votes and figure wins/losses for each team
	for _, vt := range site.Votes {
		for i := 0; i < len(vt.Choices); i++ {
			var p []string
			for j := i; j < len(vt.Choices); j++ {
				// vt.Choices[i] wins against vt.Choices[j]
				p = append(p, vt.Choices[j].Team)
				allRanks[vt.Choices[i].Team].wins[vt.Choices[j].Team]++
				allRanks[vt.Choices[j].Team].losses[vt.Choices[i].Team]++
			}
			// Now go through site.Teams for every team that isn't vt.Choices[i]
			// and isn't in 'p', mark it as a loss for the unused team
			for j := range site.Teams {
				var isUsed bool
				if site.Teams[j].UUID == vt.Choices[i].Team {
					continue
				}
				for k := range p {
					if site.Teams[j].UUID == p[k] {
						isUsed = true
					}
				}
				if !isUsed {
					allRanks[vt.Choices[i].Team].wins[site.Teams[j].UUID]++
					allRanks[site.Teams[j].UUID].losses[vt.Choices[i].Team]++
				}
			}
		}
	}
	for _, v := range allRanks {
		fmt.Println("\n" + v.tm.UUID)
		fmt.Println("  Wins:")
		for k, v := range v.wins {
			fmt.Print("    ", k, ":", v, "\n")
		}
		fmt.Println("  Losses:")
		for k, v := range v.losses {
			fmt.Print("    ", k, ":", v, "\n")
		}
	}
	return ret
}

func getInstantRunoffResult() []Team {
	var ret []Team
	return ret
}

func handleAdminVotes(w http.ResponseWriter, req *http.Request, page *pageData) {
	vars := mux.Vars(req)
	page.SubTitle = "Votes"

	type vpdVote struct {
		Timestamp string
		ClientId  string
		Choices   []Team
	}
	type votePageData struct {
		AllVotes []vpdVote
	}
	vpd := new(votePageData)
	for i := range site.Votes {
		v := new(vpdVote)
		v.Timestamp = site.Votes[i].Timestamp.Format(time.RFC3339)
		v.ClientId = site.Votes[i].ClientId
		for _, choice := range site.Votes[i].Choices {
			for _, fndTm := range site.Teams {
				if fndTm.UUID == choice.Team {
					v.Choices = append(v.Choices, fndTm)
					break
				}
			}
		}
		vpd.AllVotes = append(vpd.AllVotes, *v)
	}
	page.TemplateData = vpd
	_ = getCondorcetResult()

	switch vars["function"] {
	default:
		page.show("admin-votes.html", w)
	}
}
