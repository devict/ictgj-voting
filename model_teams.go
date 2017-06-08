package main

import (
	"github.com/pborman/uuid"
)

type Team struct {
	UUID    string
	Name    string
	Members []TeamMember
	Game    *Game
}

type TeamMember struct {
	UUID    string
	Name    string
	SlackId string
	Twitter string
	Email   string
}

func dbCreateNewTeam(nm string) error {
	var err error
	if err = db.OpenDB(); err != nil {
		return err
	}
	defer db.CloseDB()

	// Generate a UUID
	uuid := uuid.New()
	teamPath := []string{"teams", uuid}

	if err := db.MkBucketPath(teamPath); err != nil {
		return err
	}
	if err := db.SetValue(teamPath, "name", nm); err != nil {
		return err
	}
	if err := db.MkBucketPath(append(teamPath, "members")); err != nil {
		return err
	}
	gamePath := append(teamPath, "game")
	if err := db.MkBucketPath(gamePath); err != nil {
		return err
	}
	if err := db.SetValue(append(gamePath), "name", ""); err != nil {
		return err
	}
	return db.MkBucketPath(append(gamePath, "screenshots"))
}

func dbIsValidTeam(id string) bool {
	var err error
	if err = db.OpenDB(); err != nil {
		return false
	}
	defer db.CloseDB()

	teamPath := []string{"teams"}
	if teamUids, err := db.GetBucketList(teamPath); err == nil {
		for _, v := range teamUids {
			if v == id {
				return true
			}
		}
	}
	return false
}

func dbGetAllTeams() []Team {
	var ret []Team
	var err error
	if err = db.OpenDB(); err != nil {
		return ret
	}
	defer db.CloseDB()

	teamPath := []string{"teams"}
	var teamUids []string
	if teamUids, err = db.GetBucketList(teamPath); err != nil {
		return ret
	}
	for _, v := range teamUids {
		if tm := dbGetTeam(v); tm != nil {
			ret = append(ret, *tm)
		}
	}
	return ret
}

func dbGetTeam(id string) *Team {
	var err error
	if err = db.OpenDB(); err != nil {
		return nil
	}
	defer db.CloseDB()

	teamPath := []string{"teams", id}
	tm := new(Team)
	tm.UUID = id
	if tm.Name, err = db.GetValue(teamPath, "name"); err != nil {
		return nil
	}
	return tm
}

func dbGetTeamByName(nm string) *Team {
	var err error
	if err = db.OpenDB(); err != nil {
		return nil
	}
	defer db.CloseDB()

	teamPath := []string{"teams"}
	var teamUids []string
	if teamUids, err = db.GetBucketList(teamPath); err != nil {
		for _, v := range teamUids {
			var name string
			if name, err = db.GetValue(append(teamPath, v), "name"); name == nm {
				return dbGetTeam(v)
			}
		}
	}
	return nil
}

func dbUpdateTeam(id string, tm *Team) error {
	var err error
	if err = db.OpenDB(); err != nil {
		return nil
	}
	defer db.CloseDB()

	teamPath := []string{"teams", id}
	return db.SetValue(teamPath, "name", tm.Name)
}

func dbDeleteTeam(id string) error {
	var err error
	if err = db.OpenDB(); err != nil {
		return err
	}
	defer db.CloseDB()

	teamPath := []string{"teams"}
	return db.DeleteBucket(teamPath, id)
}
