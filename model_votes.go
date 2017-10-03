package main

import "time"

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

func (db *gjDatabase) getAllVotes() []Vote {
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
