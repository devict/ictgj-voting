package main

import (
	"errors"
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

	mPath []string // The path in the DB to this team
}

func NewVote(clId string, tm time.Time) (*Vote, error) {
	if clId == "" {
		return nil, errors.New("Client ID is required")
	}
	if tm.IsZero() {
		tm = time.Now()
	}

	vt := new(Vote)
	vt.mPath = []string{"jam", "votes", clId, tm.Format(time.RFC3339)}
	return vt, nil
}

func (vt *Vote) SetChoices(ch []string) error {
	// Clear any previous choices from this vote
	vt.Choices = []GameChoice{}
	for i, v := range ch {
		vt.Choices = append(vt.Choices, GameChoice{Rank: i, Team: v})
	}
	return nil
}

func (gj *Gamejam) GetVoteWithTimeString(clId, ts string) (*Vote, error) {
	timestamp, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return nil, err
	}
	return gj.GetVote(clId, timestamp)
}

func (gj *Gamejam) GetVote(clId string, ts time.Time) (*Vote, error) {
	for _, v := range gj.Votes {
		if v.ClientId == clId && v.Timestamp == ts {
			return &v, nil
		}
	}
	return nil, errors.New("Couldn't find requested vote")
}

func (gj *Gamejam) AddVote(vt *Vote) error {
	// Make sure that this isn't a duplicate
	if _, err := gj.GetVote(vt.ClientId, vt.Timestamp); err == nil {
		return errors.New("Duplicate Vote")
	}
	gj.Votes = append(gj.Votes, *vt)
	return nil
}

/**
 * DB Functions
 * These are generally just called when the app starts up or when the periodic 'save' runs
 */

// LoadAllVotes loads all votes for the jam out of the database
func (gj *Gamejam) LoadAllVotes() []Vote {
	var err error
	var ret []Vote
	if err = gj.m.openDB(); err != nil {
		return ret
	}
	defer gj.m.closeDB()

	votesPath := []string{"jam", "votes"}
	var cliUUIDs []string
	if cliUUIDs, err = gj.m.bolt.GetBucketList(votesPath); err != nil {
		return ret
	}
	for _, cId := range cliUUIDs {
		vtsPth := append(votesPath, cId)
		var times []string
		if times, err = gj.m.bolt.GetBucketList(vtsPth); err != nil {
			// Error reading this bucket, move on to the next
			continue
		}
		for _, t := range times {
			if vt, err := gj.LoadVote(cId, t); err == nil {
				ret = append(ret, *vt)
			}
		}
	}
	return ret
}

// Load a vote from the DB and return it
func (gj *Gamejam) LoadVote(clientId, t string) (*Vote, error) {
	var tm time.Time
	var err error
	if tm, err = time.Parse(time.RFC3339, t); err != nil {
		return nil, errors.New("Error loading vote: " + err.Error())
	}
	vt, err := NewVote(clientId, tm)
	if err != nil {
		return nil, errors.New("Error creating vote: " + err.Error())
	}
	var choices []string
	if choices, err = m.bolt.GetKeyList(vt.mPath); err != nil {
		return nil, errors.New("Error creating vote: " + err.Error())
	}
	for _, v := range choices {
		ch := new(GameChoice)
		var rank int
		if rank, err = strconv.Atoi(v); err == nil {
			ch.Rank = rank
			ch.Team, _ = m.bolt.GetValue(vt.mPath, v)
			vt.Choices = append(vt.Choices, *ch)
		}
	}
	return vt, nil
}

func (gj *Gamejam) SaveVote(vt *Vote) error {
	var err error
	if err = gj.m.openDB(); err != nil {
		return err
	}
	defer gj.m.closeDB()

	for _, v := range vt.Choices {
		m.bolt.SetValue(vt.mPath, strconv.Itoa(v.Rank), v.Team)
	}
	return nil
}
