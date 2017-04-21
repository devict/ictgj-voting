package main

import (
	"fmt"

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

	var currJam string
	if currJam, err = dbGetCurrentJam(); err != nil {
		return err
	}

	// Generate a UUID
	uuid := uuid.New()
	teamPath := []string{"jams", currJam, "teams", uuid}

	if err := db.MkBucketPath(teamPath); err != nil {
		fmt.Println("Error at 39: " + uuid)
		return err
	}
	if err := db.SetValue(teamPath, "name", nm); err != nil {
		fmt.Println("Error at 43")
		return err
	}
	if err := db.MkBucketPath(append(teamPath, "members")); err != nil {
		fmt.Println("Error at 47")
		return err
	}
	gamePath := append(teamPath, "game")
	if err := db.MkBucketPath(gamePath); err != nil {
		fmt.Println("Error at 52")
		return err
	}
	if err := db.SetValue(append(gamePath), "name", ""); err != nil {
		fmt.Println("Error at 56")
		return err
	}
	return db.MkBucketPath(append(gamePath, "screenshots"))
}

func dbIsValidTeam(nm string) bool {
	var err error
	var currJam string
	if err = db.OpenDB(); err != nil {
		return false
	}
	defer db.CloseDB()

	if currJam, err = dbGetCurrentJam(); err != nil {
		return false
	}
	teamPath := []string{"jams", currJam, "teams"}
	if teamUids, err := db.GetBucketList(teamPath); err == nil {
		for _, v := range teamUids {
			if tstName, err := db.GetValue(append(teamPath, v), "name"); err == nil {
				if tstName == nm {
					return true
				}
			}
		}
	}
	return false
}

func dbGetAllTeams() []Team {
	var ret []Team
	var err error
	var currJam string
	if err = db.OpenDB(); err != nil {
		return ret
	}
	defer db.CloseDB()

	if currJam, err = dbGetCurrentJam(); err != nil {
		return ret
	}
	teamPath := []string{"jams", currJam, "teams"}
	if teamUids, err := db.GetBucketList(teamPath); err != nil {
		for _, v := range teamUids {
			if tm := dbGetTeam(v); tm != nil {
				ret = append(ret, *tm)
			}
		}
	}
	return ret
}

func dbGetTeam(id string) *Team {
	var err error
	var currJam string
	if err = db.OpenDB(); err != nil {
		return nil
	}
	defer db.CloseDB()

	if currJam, err = dbGetCurrentJam(); err != nil {
		return nil
	}
	teamPath := []string{"jams", currJam, "teams", id}
	tm := new(Team)
	if tm.Name, err = db.GetValue(teamPath, "name"); err != nil {
		return nil
	}
	return tm
}

func dbGetTeamByName(nm string) *Team {
	var err error
	var currJam string
	if err = db.OpenDB(); err != nil {
		return nil
	}
	defer db.CloseDB()

	if currJam, err = dbGetCurrentJam(); err != nil {
		return nil
	}
	teamPath := []string{"jams", currJam, "teams"}
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

func dbDeleteTeam(id string) error {
	var err error
	var currJam string
	if err = db.OpenDB(); err != nil {
		return err
	}
	defer db.CloseDB()

	if currJam, err = dbGetCurrentJam(); err != nil {
		return err
	}
	teamPath := []string{"jams", currJam, "teams"}
	return db.DeleteBucket(teamPath, id)
}
