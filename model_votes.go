package main

import "time"

type Vote struct {
	Timestamp time.Time
	ClientId  string
}

func dbGetAllVotes() []Vote {
	var ret []Vote
	var err error
	if err = db.OpenDB(); err != nil {
		return ret
	}
	defer db.CloseDB()

	votesBkt := []string{"votes"}
	return ret
}

func dbGetVote(clientId string, timestamp time.Time) *Vote {
	var err error
	if err = db.OpenDB(); err != nil {
		return nil
	}
	defer db.CloseDB()

	vt := new(Vote)

	return vt
}

func dbSaveVote(clientId string, timestamp time.Time, votes []string) error {
	var err error
	if err = db.OpenDB(); err != nil {
		return nil
	}
	defer db.CloseDB()

	votesBkt := []string{"votes", clientId}

	return err
}
