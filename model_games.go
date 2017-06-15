package main

import (
	"errors"

	"github.com/pborman/uuid"
)

type Game struct {
	Name        string
	TeamId      string
	Description string
	Screenshots []Screenshot
}

type Screenshot struct {
	Description string
	Image       string
}

func dbUpdateTeamGame(teamId, name, desc string) error {
	var err error
	if err = openDatabase(); err != nil {
		return err
	}
	defer closeDatabase()

	// Make sure the team is valid
	tm := dbGetTeam(teamId)
	if tm == nil {
		return errors.New("Invalid team")
	}
	gamePath := []string{"teams", teamId, "game"}

	if err := db.MkBucketPath(gamePath); err != nil {
		return err
	}
	if name == "" {
		name = tm.Name + "'s Game"
	}
	if err := db.SetValue(gamePath, "name", name); err != nil {
		return err
	}
	if err := db.SetValue(gamePath, "description", desc); err != nil {
		return err
	}
	if err := db.MkBucketPath(append(gamePath, "screenshots")); err != nil {
		return err
	}

	return err
}

func dbGetAllGames() []Game {
	var ret []Game
	tms := dbGetAllTeams()
	for i := range tms {
		ret = append(ret, *dbGetTeamGame(tms[i].UUID))
	}
	return ret
}

func dbGetTeamGame(teamId string) *Game {
	var err error
	if err = openDatabase(); err != nil {
		return nil
	}
	defer closeDatabase()

	gamePath := []string{"teams", teamId, "game"}
	gm := new(Game)
	if gm.Name, err = db.GetValue(gamePath, "name"); err != nil {
		gm.Name = ""
	}
	gm.TeamId = teamId
	if gm.Description, err = db.GetValue(gamePath, "description"); err != nil {
		gm.Description = ""
	}
	gm.Screenshots = dbGetTeamGameScreenshots(teamId)
	return gm
}

// Screenshots are saved as base64 encoded pngs
func dbGetTeamGameScreenshots(teamId string) []Screenshot {
	var ret []Screenshot
	var err error
	ssPath := []string{"teams", teamId, "game", "screenshots"}
	var ssIds []string
	if ssIds, err = db.GetBucketList(ssPath); err != nil {
		return ret
	}
	for _, v := range ssIds {
		if ss := dbGetTeamGameScreenshot(teamId, v); ss != nil {
			ret = append(ret, *ss)
		}
	}
	return ret
}

func dbGetTeamGameScreenshot(teamId, ssId string) *Screenshot {
	var err error
	ssPath := []string{"teams", teamId, "game", "screenshots", ssId}
	ret := new(Screenshot)
	if ret.Description, err = db.GetValue(ssPath, "description"); err != nil {
		return nil
	}
	if ret.Image, err = db.GetValue(ssPath, "image"); err != nil {
		return nil
	}
	return ret
}

func dbSaveTeamGameScreenshot(teamId string, ss *Screenshot) error {
	var err error
	if err = openDatabase(); err != nil {
		return nil
	}
	defer closeDatabase()

	ssPath := []string{"teams", teamId, "game", "screenshots"}
	// Generate a UUID for this screenshot
	uuid := uuid.New()
	ssPath = append(ssPath, uuid)
	if err := db.MkBucketPath(ssPath); err != nil {
		return err
	}
	if err := db.SetValue(ssPath, "description", ss.Description); err != nil {
		return err
	}
	if err := db.SetValue(ssPath, "image", ss.Image); err != nil {
		return err
	}
	return nil
}

func dbDeleteTeamGameScreenshot(teamId, ssId string) error {
	var err error
	if err = openDatabase(); err != nil {
		return nil
	}
	defer closeDatabase()

	ssPath := []string{"teams", teamId, "game", "screenshots"}
	return db.DeleteBucket(ssPath, ssId)
}
