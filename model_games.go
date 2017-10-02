package main

import "errors"

type Game struct {
	Name        string
	TeamId      string
	Link        string
	Description string
	Screenshots []Screenshot
}

type Screenshot struct {
	UUID        string
	Description string
	Image       string
	Thumbnail   string
	Filetype    string
}

// Create a new game object, must have a valid team id
func newGame(tmId string) *Game {
	var err error
	if err = db.open(); err != nil {
		return nil
	}
	defer db.close()

	tm := db.getTeam(tmId)
	if tm == nil {
		return nil
	}
	return &Game{TeamId: tmId}
}

func (gm *Game) save() error {
	var err error
	if err = db.open(); err != nil {
		return err
	}
	defer db.close()

	tm := db.getTeam(gm.TeamId)
	if tm == nil {
		return errors.New("Invalid Team: " + gm.TeamId)
	}
	gamePath := []string{"teams", gm.TeamId, "game"}
	if err := db.bolt.MkBucketPath(gamePath); err != nil {
		return err
	}

	if gm.Name == "" {
		gm.Name = tm.Name + "'s Game"
	}
	if err := db.bolt.SetValue(gamePath, "name", gm.Name); err != nil {
		return err
	}
	if err := db.bolt.SetValue(gamePath, "link", gm.Link); err != nil {
		return err
	}
	if err := db.bolt.SetValue(gamePath, "description", gm.Description); err != nil {
		return err
	}
	if err := db.bolt.MkBucketPath(append(gamePath, "screenshots")); err != nil {
		return err
	}

	return err
}

func (db *gjDatabase) getAllGames() []Game {
	var ret []Game
	tms := db.getAllTeams()
	for i := range tms {
		ret = append(ret, *tms[i].getGame())
	}
	return ret
}
