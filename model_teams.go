package main

import (
	"errors"

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

func dbEditTeamGame(teamid, name string) error {
	var err error
	if err = db.OpenDB(); err != nil {
		return err
	}
	defer db.CloseDB()

	gamePath := []string{"teams", teamid, "game"}
	return db.SetValue(gamePath, "name", name)
}

func dbGetTeamMembers(teamid string) ([]TeamMember, error) {
	var ret []TeamMember
	var err error
	if err = db.OpenDB(); err != nil {
		return ret, nil
	}
	defer db.CloseDB()

	teamMbrPath := []string{"teams", teamid, "members"}
	var memberUuids []string
	if memberUuids, err = db.GetBucketList(teamMbrPath); err != nil {
		for _, v := range memberUuids {
			var mbr *TeamMember
			if mbr, err = dbGetTeamMember(teamid, v); err != nil {
				ret = append(ret, *mbr)
			}
		}
	}
	return ret, nil
}

func dbGetTeamMember(teamid, mbrid string) (*TeamMember, error) {
	var err error
	if err = db.OpenDB(); err != nil {
		return nil, err
	}
	defer db.CloseDB()

	teamMbrPath := []string{"teams", teamid, "members", mbrid}
	var memberUuids []string
	if memberUuids, err = db.GetBucketList(teamMbrPath); err != nil {
		for _, v := range memberUuids {
			mbr := new(TeamMember)
			mbr.UUID = v
			if mbr.Name, err = db.GetValue(append(teamMbrPath, v), "name"); err != nil {
				return nil, err
			}
			if mbr.SlackId, err = db.GetValue(append(teamMbrPath, v), "slackid"); err != nil {
				return nil, err
			}
			if mbr.Twitter, err = db.GetValue(append(teamMbrPath, v), "twitter"); err != nil {
				return nil, err
			}
			if mbr.Email, err = db.GetValue(append(teamMbrPath, v), "email"); err != nil {
				return nil, err
			}
			return mbr, err
		}
	}
	return nil, errors.New("Couldn't find team member")
}
