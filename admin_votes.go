package main

import (
	"errors"
	"net/http"
	"strings"
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
	for i := 0; i < len(m.jam.Teams); i++ {
		for j := i + 1; j < len(m.jam.Teams); j++ {
			// For each pairing find a winner
			winner, pct, _ := findWinnerBetweenTeams(&m.jam.Teams[i], &m.jam.Teams[j])
			newPair := new(teamPair)
			if winner != nil {
				newPair.winner = winner
				if winner.UUID == m.jam.Teams[i].UUID {
					newPair.loser = &m.jam.Teams[j]
				} else {
					newPair.loser = &m.jam.Teams[i]
				}
				newPair.majority = pct
			} else {
				newPair.winner = &m.jam.Teams[i]
				newPair.loser = &m.jam.Teams[j]
				newPair.majority = 50
			}
			allPairs = append(allPairs, *newPair)
		}
	}
	// initialize map of team wins
	teamWins := make(map[string]int)
	for i := range m.jam.Teams {
		teamWins[m.jam.Teams[i].UUID] = 0
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
			tm, _ := m.jam.GetTeamById(rankedWins[topWins][i])
			if tm != nil {
				nR.Teams = append(nR.Teams, *tm)
			}
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
	for _, v := range m.jam.Votes {
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
		Timestamp   string
		ClientId    string
		Choices     []Team
		VoterStatus string
		Discovery   string
	}
	type votePageData struct {
		AllVotes      []vpdVote
		Results       []Ranking
		VoterStatuses map[string]int
	}
	vpd := new(votePageData)
	vpd.VoterStatuses = make(map[string]int)
	now := time.Now()
	dayThresh := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	for i := range m.jam.Votes {
		v := new(vpdVote)
		if m.jam.Votes[i].Timestamp.Before(dayThresh) {
			v.Timestamp = m.jam.Votes[i].Timestamp.Format("Jan _2 15:04")
		} else {
			v.Timestamp = m.jam.Votes[i].Timestamp.Format(time.Kitchen)
		}
		v.ClientId = m.jam.Votes[i].ClientId
		for _, choice := range m.jam.Votes[i].Choices {
			for _, fndTm := range m.jam.Teams {
				if fndTm.UUID == choice.Team {
					v.Choices = append(v.Choices, fndTm)
					break
				}
			}
		}
		v.VoterStatus = m.jam.Votes[i].VoterStatus
		if strings.TrimSpace(v.VoterStatus) != "" {
			vpd.VoterStatuses[v.VoterStatus]++
		}
		v.Discovery = m.jam.Votes[i].Discovery
		vpd.AllVotes = append(vpd.AllVotes, *v)
	}
	vpd.Results = getCondorcetResult()
	page.TemplateData = vpd

	switch vars["function"] {
	default:
		page.show("admin-votes.html", w)
	}
}
