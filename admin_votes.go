package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Ranking struct {
	Rank  int
	Teams []Team
}

// getCondorcetResult returns the ranking of teams based on the condorcet method
// https://en.wikipedia.org/wiki/Condorcet_method
func getCondorcetResult() []Ranking {
	type teamPair struct {
		winner   *Team
		loser    *Team
		majority float32
	}
	var allPairs []teamPair
	var ret []Ranking
	for i := 0; i < len(site.Teams); i++ {
		for j := i + 1; j < len(site.Teams); j++ {
			// For each pairing find a winner
			winner, pct, _ := findWinnerBetweenTeams(&site.Teams[i], &site.Teams[j])
			newPair := new(teamPair)
			if winner != nil {
				newPair.winner = winner
				if winner.UUID == site.Teams[i].UUID {
					newPair.loser = &site.Teams[j]
				} else {
					newPair.loser = &site.Teams[i]
				}
				newPair.majority = pct
			} else {
				newPair.winner = &site.Teams[i]
				newPair.loser = &site.Teams[j]
				newPair.majority = 50
			}
			allPairs = append(allPairs, *newPair)
		}
	}
	// initialize map of team wins
	teamWins := make(map[string]int)
	for i := range site.Teams {
		teamWins[site.Teams[i].UUID] = 0
	}
	// Figure out how many wins each team has
	for i := range allPairs {
		if allPairs[i].majority != 50 {
			teamWins[allPairs[i].winner.UUID]++
		}
	}

	// Rank them by wins
	rankedWins := make(map[int][]string)
	for k, v := range teamWins {
		rankedWins[v] = append(rankedWins[v], k)
	}
	currRank := 1
	for len(rankedWins) > 0 {
		topWins := 0
		for k, _ := range rankedWins {
			if k > topWins {
				topWins = k
			}
		}
		nR := new(Ranking)
		nR.Rank = currRank
		for i := range rankedWins[topWins] {
			nR.Teams = append(nR.Teams, *site.getTeamByUUID(rankedWins[topWins][i]))
		}
		ret = append(ret, *nR)
		delete(rankedWins, topWins)
		currRank++
	}
	return ret
}

// This is a helper function for calculating results
func uuidIsInRankingSlice(uuid string, sl []Ranking) bool {
	for _, v := range sl {
		for i := range v.Teams {
			if v.Teams[i].UUID == uuid {
				return true
			}
		}
	}
	return false
}

// findWinnerBetweenTeams returns the team that got the most votes
// and the percentage of votes they received
// or an error if a winner couldn't be determined.
func findWinnerBetweenTeams(tm1, tm2 *Team) (*Team, float32, error) {
	// tally gets incremented for a tm1 win, decremented for a tm2 win
	var tm1votes, tm2votes float32
	for _, v := range site.Votes {
		for _, chc := range v.Choices {
			if chc.Team == tm1.UUID {
				tm1votes++
				break
			} else if chc.Team == tm2.UUID {
				tm2votes++
				break
			}
		}
	}
	ttlVotes := tm1votes + tm2votes
	if tm1votes > tm2votes {
		return tm1, 100 * (tm1votes / ttlVotes), nil
	} else if tm1votes < tm2votes {
		return tm2, 100 * (tm2votes / ttlVotes), nil
	}
	return nil, 50, errors.New("Unable to determine a winner")
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
		Results  []Ranking
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
	vpd.Results = getCondorcetResult()
	page.TemplateData = vpd

	switch vars["function"] {
	default:
		page.show("admin-votes.html", w)
	}
}
