package main

import "errors"

/**
 * Game
 * A team's game, including links, description, and screenshots
 */
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

// Load a team's game from the DB and return it
func (gj *Gamejam) LoadTeamGame(tmId string) *Game {
	var err error
	if err = gj.m.openDB(); err != nil {
		return nil
	}
	defer gj.m.closeDB()

	gamePath := []string{"jam", "teams", tmId, "game"}
	gm := new(Game)
	gm.TeamId = tm.UUID
	if gm.Name, err = gj.m.bolt.GetValue(gamePath, "name"); err != nil {
		gm.Name = ""
	}
	if gm.Description, err = gj.m.bolt.GetValue(gamePath, "description"); err != nil {
		gm.Description = ""
	}
	if gm.Link, err = gj.m.bolt.GetValue(gamePath, "link"); err != nil {
		gm.Link = ""
	}
	// Now get the game screenshots
	gm.Screenshots = gj.LoadTeamGameScreenshots(tmId)

	return &gm
}

func (gj *Gamejam) LoadTeamGameScreenshots(tmId string) []Screenshot {
	var err error
	if err = gj.m.openDB(); err != nil {
		return nil
	}
	defer gj.m.closeDB()

	var ret []Screenshot
	ssBktPath := []string{"jam", "teams", tmId, "game", "screenshots"}
	var ssIds []string
	ssIds, _ = gj.m.bolt.GetBucketList(ssBktPath)
	for _, v := range ssIds {
		ssLd := gj.LoadTeamGameScreenshot(tmId, v)
		if ssLd != nil {
			ret = append(ret, ssLd)
		}
	}
	return ret
}

func (gj *Gamejam) LoadTeamGameScreenshot(tmId, ssId string) *Screenshot {
	var err error
	if err = gj.m.openDB(); err != nil {
		return nil
	}
	defer gj.m.closeDB()

	var ret []Screenshot
	ssPath := []string{"jam", "teams", tmId, "game", "screenshots", ssId}
	ret := new(Screenshot)
	ret.UUID = ssId
	if ret.Description, err = gj.m.bolt.GetValue(ssPath, "description"); err != nil {
		return nil
	}
	if ret.Image, err = gj.m.bolt.GetValue(ssPath, "image"); err != nil {
		return nil
	}
	if ret.Thumbnail, err = gj.m.bolt.GetValue(ssPath, "thumbnail"); err != nil {
		return nil
	}
	if ret.Thumbnail == "" {
		ret.Thumbnail = ret.Image
	}
	if ret.Filetype, err = gj.m.bolt.GetValue(ssPath, "filetype"); err != nil {
		return nil
	}
	return ret
}

/**
 * OLD FUNCTIONS
 */

// Create a new game object, must have a valid team id
func (db *currJamDb) newGame(tmId string) *Game {
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

func (db *currJamDb) getAllGames() []Game {
	var ret []Game
	tms := db.getAllTeams()
	for i := range tms {
		ret = append(ret, *tms[i].getGame())
	}
	return ret
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
