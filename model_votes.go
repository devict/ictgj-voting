package main

import (
	"strconv"
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

// LoadAllVotes loads all votes for the jam out of the database
func (gj *Gamejam) LoadAllVotes() []Vote {
	var ret []Vote
	if err := gj.m.openDB(); err != nil {
		return err
	}
	defer gj.m.closeDB()

	votesPath := []string{"jam", "votes"}
	if cliUUIDs, err = m.bolt.GetBucketList(votesPath); err != nil {
		return ret
	}
	for _, cId := range cliUUIDs {
		vtsPth := append(votesPath, cId)
		if times, err := m.bolt.GetBucketList(vtsPth); err != nil {
			// Error reading this bucket, move on to the next
			continue
		}
		for _, t := range times {
			vt := gj.LoadVote(cId, t)
			if vt != nil {
				ret = append(ret, vt)
			}
		}
	}
	return ret
}

// Load a vote from the DB and return it
func (gj *Gamejam) LoadVote(clientId, tm string) *Vote {
	var tm time.Time
	if tm, err = time.Parse(time.RFC3339, t); err != nil {
		return nil
	}
	vt := new(Vote)
	vt.Timestamp = tm
	vt.ClientId = cId
	vtPth := append(vtsPth, t)
	var choices []string
	if choices, err = m.bolt.GetKeyList(vtPth); err != nil {
		return nil
	}
	for _, v := range choices {
		ch := new(GameChoices)
		var rank int
		if rank, err = strconv.Atoi(v); err == nil {
			ch.Rank = rank
			ch.Team, _ = m.bolt.GetValue(vtPth, v)
			vt.Choices = append(vt.Choices, *ch)
		}
	}
	return &vt
}

/**
 * OLD FUNCTIONS
 */
func (db *currJamDb) getAllVotes() []Vote {
	var ret []Vote
	var err error
	if err = db.open(); err != nil {
		return ret
	}
	defer db.close()

	clients := db.getAllClients()
	for _, cl := range clients {
		ret = append(ret, cl.getVotes()...)
	}
	return ret
}
