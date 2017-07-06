package main

import (
	"strconv"
	"strings"
	"time"
)

// A Choice is a ranking of a game in a vote
type GameChoice struct {
	Team string // UUID of team
	Rank int
}

// A Vote is a collection of game rankings
type Vote struct {
	Timestamp time.Time
	ClientId  string // UUID of client
	Choices   []GameChoice
}

func dbGetAllVotes() []Vote {
	var ret []Vote
	var err error
	if err = db.OpenDB(); err != nil {
		return ret
	}
	defer db.CloseDB()

	votesBkt := []string{"votes"}
	var clients []string
	if clients, err = db.GetBucketList(votesBkt); err != nil {
		// Couldn't get the list of clients
		return ret
	}
	for _, clid := range clients {
		ret = append(ret, dbGetClientVotes(clid)...)
	}
	return ret
}

func dbGetClientVotes(clientId string) []Vote {
	var ret []Vote
	var err error
	if err = db.OpenDB(); err != nil {
		return ret
	}
	defer db.CloseDB()

	var times []string
	votesBkt := []string{"votes", clientId}
	if times, err = db.GetBucketList(votesBkt); err != nil {
		return ret
	}
	for _, t := range times {
		var tm time.Time
		if tm, err = time.Parse(time.RFC3339, t); err == nil {
			var vt *Vote
			if vt, err = dbGetVote(clientId, tm); err == nil {
				ret = append(ret, *vt)
			}
		}
	}
	return ret
}

func dbGetVote(clientId string, timestamp time.Time) (*Vote, error) {
	var err error
	if err = db.OpenDB(); err != nil {
		return nil, err
	}
	defer db.CloseDB()

	vt := new(Vote)
	vt.Timestamp = timestamp
	vt.ClientId = clientId
	votesBkt := []string{"votes", clientId, timestamp.Format(time.RFC3339)}
	var choices []string
	if choices, err = db.GetKeyList(votesBkt); err != nil {
		// Couldn't find the vote...
		return nil, err
	}
	for _, v := range choices {
		ch := new(GameChoice)
		var rank int

		if rank, err = strconv.Atoi(v); err == nil {
			ch.Rank = rank
			ch.Team, err = db.GetValue(votesBkt, v)
			vt.Choices = append(vt.Choices, *ch)
		}
	}
	return vt, nil
}

func dbSaveVote(clientId string, timestamp time.Time, votes []string) error {
	var err error
	if err = db.OpenDB(); err != nil {
		return nil
	}
	defer db.CloseDB()
	// Make sure we don't clobber a duplicate vote
	votesBkt := []string{"votes", clientId, timestamp.Format(time.RFC3339)}
	for i := range votes {
		if strings.TrimSpace(votes[i]) != "" {
			db.SetValue(votesBkt, strconv.Itoa(i), votes[i])
		}
	}
	return err
}
