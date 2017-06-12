package main

import (
	"errors"
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
	if err = openDatabase(); err != nil {
		return err
	}
	defer closeDatabase()

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
	if err = openDatabase(); err != nil {
		return false
	}
	defer closeDatabase()

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
	if err = openDatabase(); err != nil {
		return ret
	}
	defer closeDatabase()

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
	if err = openDatabase(); err != nil {
		return nil
	}
	defer closeDatabase()

	teamPath := []string{"teams", id}
	tm := new(Team)
	tm.UUID = id
	if tm.Name, err = db.GetValue(teamPath, "name"); err != nil {
		return nil
	}
	tm.Members, _ = dbGetTeamMembers(id)
	return tm
}

func dbGetTeamByName(nm string) *Team {
	var err error
	if err = openDatabase(); err != nil {
		return nil
	}
	defer closeDatabase()

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
	if err = openDatabase(); err != nil {
		return err
	}
	defer closeDatabase()

	teamPath := []string{"teams", id}
	return db.SetValue(teamPath, "name", tm.Name)
}

func dbDeleteTeam(id string) error {
	var err error
	if err = openDatabase(); err != nil {
		return err
	}
	defer closeDatabase()

	teamPath := []string{"teams"}
	return db.DeleteBucket(teamPath, id)
}

func dbEditTeamGame(teamid, name string) error {
	var err error
	if err = openDatabase(); err != nil {
		return err
	}
	defer closeDatabase()

	gamePath := []string{"teams", teamid, "game"}
	return db.SetValue(gamePath, "name", name)
}

func dbAddTeamMember(teamid, mbrName, mbrEmail, mbrSlack, mbrTwitter string) error {
	// First check if this member already exists on this team
	mbrs, _ := dbGetTeamMembers(teamid)
	if len(mbrs) > 0 {
		for i := range mbrs {
			if mbrs[i].Name == mbrName {
				return dbUpdateTeamMember(teamid, mbrs[i].UUID, mbrName, mbrEmail, mbrSlack, mbrTwitter)
			}
		}
	}
	// It's really an add
	mbrId := uuid.New()
	return dbUpdateTeamMember(teamid, mbrId, mbrName, mbrEmail, mbrSlack, mbrTwitter)
}

func dbUpdateTeamMember(teamid, mbrId, mbrName, mbrEmail, mbrSlack, mbrTwitter string) error {
	var err error
	if err = openDatabase(); err != nil {
		return err
	}
	defer closeDatabase()

	mbrPath := []string{"teams", teamid, "members", mbrId}
	if db.SetValue(mbrPath, "name", mbrName) != nil {
		return err
	}
	if db.SetValue(mbrPath, "slackid", mbrSlack) != nil {
		return err
	}
	if db.SetValue(mbrPath, "twitter", mbrTwitter) != nil {
		return err
	}
	if db.SetValue(mbrPath, "email", mbrEmail) != nil {
		return err
	}
	return nil
}

func dbGetTeamMembers(teamid string) ([]TeamMember, error) {
	var ret []TeamMember
	var err error
	if err = openDatabase(); err != nil {
		return ret, err
	}
	defer closeDatabase()

	teamPath := []string{"teams", teamid, "members"}
	var memberUuids []string
	if memberUuids, err = db.GetBucketList(teamPath); err == nil {
		for _, v := range memberUuids {
			var mbr *TeamMember
			if mbr, err = dbGetTeamMember(teamid, v); err == nil {
				fmt.Println("Finding Team Members", teamid, mbr.Name)
				ret = append(ret, *mbr)
			}
		}
	} else {
		fmt.Println(err.Error())
	}
	return ret, nil
}

func dbGetTeamMember(teamid, mbrid string) (*TeamMember, error) {
	var err error
	if err = openDatabase(); err != nil {
		return nil, err
	}
	defer closeDatabase()

	mbr := new(TeamMember)
	teamMbrPath := []string{"teams", teamid, "members", mbrid}
	mbr.UUID = mbrid
	if mbr.Name, err = db.GetValue(teamMbrPath, "name"); err != nil {
		return nil, err
	}
	if mbr.SlackId, err = db.GetValue(teamMbrPath, "slackid"); err != nil {
		return nil, err
	}
	if mbr.Twitter, err = db.GetValue(teamMbrPath, "twitter"); err != nil {
		return nil, err
	}
	if mbr.Email, err = db.GetValue(teamMbrPath, "email"); err != nil {
		return nil, err
	}
	return mbr, err
}

// This function returns the team for a specific member
func dbGetMembersTeam(mbrid string) (*Team, error) {
	var err error
	if err = openDatabase(); err != nil {
		return nil, err
	}
	defer closeDatabase()

	teams := dbGetAllTeams()
	for i := range teams {
		var tmMbrs []TeamMember
		tmMbrs, err = dbGetTeamMembers(teams[i].UUID)
		if err == nil {
			for j := range tmMbrs {
				if tmMbrs[j].UUID == mbrid {
					return &teams[i], nil
				}
			}
		}
	}
	return nil, errors.New("Unable to find team member")
}

// This function searches all teams for a member with the given name
func dbGetTeamMembersByName(mbrName string) ([]TeamMember, error) {
	var ret []TeamMember
	var err error
	if err = openDatabase(); err != nil {
		return ret, err
	}
	defer closeDatabase()

	teams := dbGetAllTeams()
	for i := range teams {
		var tmMbrs []TeamMember
		tmMbrs, err = dbGetTeamMembers(teams[i].UUID)
		if err == nil {
			for j := range tmMbrs {
				if tmMbrs[j].Name == mbrName {
					ret = append(ret, tmMbrs[j])
				}
			}
		}
	}
	if len(ret) == 0 {
		return nil, errors.New("Couldn't find any members with the requested name")
	}
	return ret, nil
}

func dbDeleteTeamMember(teamId, mbrId string) error {
	var err error
	if err = openDatabase(); err != nil {
		return err
	}
	defer closeDatabase()

	teamPath := []string{"teams", teamId, "members"}
	return db.DeleteBucket(teamPath, mbrId)
}
