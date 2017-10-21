package main

import (
	"errors"
	"fmt"

	"github.com/pborman/uuid"
)

/**
 * Team
 */
type Team struct {
	UUID    string
	Name    string
	Members []TeamMember
	Game    *Game

	mPath []string // The path in the DB to this team
}

// Create a team
func NewTeam(id string) *Team {
	if id == "" {
		id = uuid.New()
	}
	// Create an emtpy game for the team
	gm, _ := NewGame(id)
	return &Team{
		UUID:  id,
		Game:  gm,
		mPath: []string{"jam", "teams", id},
	}
}

func (gj *Gamejam) GetTeamById(id string) (*Team, error) {
	for i := range gj.Teams {
		if gj.Teams[i].UUID == id {
			return &gj.Teams[i], nil
		}
	}
	return nil, errors.New("Invalid Team Id given")
}

type TeamMember struct {
	UUID    string
	Name    string
	SlackId string
	Twitter string
	Email   string

	mPath []string // The path in the DB to this team member
}

// Create a new team member
func NewTeamMember(tmId, uId string) (*TeamMember, error) {
	if tmId == "" {
		return nil, errors.New("Team ID is required")
	}
	if uId == "" {
		uId = uuid.New()
	}
	return &TeamMember{
		UUID:  uId,
		mPath: []string{"jam", "teams", tmId, "members", uId},
	}, nil
}

// AddTeamMember adds a new team member
func (tm *Team) AddTeamMember(mbr *TeamMember) error {
	lkup, _ := tm.GetTeamMemberById(mbr.UUID)
	if lkup != nil {
		return errors.New("A Team Member with that Id already exists")
	}
	tm.Members = append(tm.Members, *mbr)
	return nil
}

// GetTeamMemberById returns a member with the given uuid
// or an error if it couldn't find it
func (tm *Team) GetTeamMemberById(uuid string) (*TeamMember, error) {
	for i := range tm.Members {
		if tm.Members[i].UUID == uuid {
			return &tm.Members[i], nil
		}
	}
	return nil, errors.New("Invalid Team Member Id given")
}

func (tm *Team) RemoveTeamMemberById(id string) error {
	idx := -1
	for i := range tm.Members {
		if tm.Members[i].UUID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return errors.New("Invalid Team Member ID given")
	}
	tm.Members = append(tm.Members[:idx], tm.Members[idx+1:]...)
	return nil
}

/**
 * DB Functions
 * These are generally just called when the app starts up, or when the periodic 'save' runs
 */

// LoadAllTeams loads all teams for the jam out of the database
func (gj *Gamejam) LoadAllTeams() []Team {
	var err error
	var ret []Team
	if err = gj.m.openDB(); err != nil {
		return ret
	}
	defer gj.m.closeDB()

	var tmUUIDs []string
	tmsPath := append(gj.mPath, "teams")
	if tmUUIDs, err = gj.m.bolt.GetBucketList(tmsPath); err != nil {
		fmt.Println(err.Error())
		return ret
	}
	for _, v := range tmUUIDs {
		tm, _ := gj.LoadTeam(v)
		if tm != nil {
			ret = append(ret, *tm)
		}
	}
	return ret
}

// Load a team out of the database
func (gj *Gamejam) LoadTeam(uuid string) (*Team, error) {
	var err error
	if err = gj.m.openDB(); err != nil {
		return nil, err
	}
	defer gj.m.closeDB()

	// Team Data
	tm := NewTeam(uuid)
	if tm.Name, err = gj.m.bolt.GetValue(tm.mPath, "name"); err != nil {
		return nil, errors.New("Error loading team: " + err.Error())
	}

	// Team Members
	tm.Members = gj.LoadTeamMembers(uuid)

	// Team Game
	if tm.Game, err = gj.LoadTeamGame(uuid); err != nil {
		return nil, errors.New("Error loading team game: " + err.Error())
	}

	return tm, nil
}

// Load the members of a team from the DB and return them
func (gj *Gamejam) LoadTeamMembers(tmId string) []TeamMember {
	var err error
	var ret []TeamMember
	if err = gj.m.openDB(); err != nil {
		return ret
	}
	defer gj.m.closeDB()

	// Team Members
	var memberUuids []string
	tm := NewTeam(tmId)
	mbrsPath := append(tm.mPath, "members")
	if memberUuids, err = gj.m.bolt.GetBucketList(mbrsPath); err == nil {
		for _, v := range memberUuids {
			mbr, _ := gj.LoadTeamMember(tmId, v)
			if mbr != nil {
				ret = append(ret, *mbr)
			}
		}
	}
	return ret
}

// Load a team member from the DB and return it
func (gj *Gamejam) LoadTeamMember(tmId, mbrId string) (*TeamMember, error) {
	var err error
	if err = gj.m.openDB(); err != nil {
		return nil, err
	}
	defer gj.m.closeDB()

	mbr, err := NewTeamMember(tmId, mbrId)
	if err != nil {
		return nil, errors.New("Error loading team member: " + err.Error())
	}
	// Name is the only required field
	if mbr.Name, err = gj.m.bolt.GetValue(mbr.mPath, "name"); err != nil {
		return nil, errors.New("Error loading team member: " + err.Error())
	}
	if mbr.SlackId, err = gj.m.bolt.GetValue(mbr.mPath, "slackid"); err != nil {
		mbr.SlackId = ""
	}
	if mbr.Twitter, err = gj.m.bolt.GetValue(mbr.mPath, "twitter"); err != nil {
		mbr.Twitter = ""
	}
	if mbr.Email, err = gj.m.bolt.GetValue(mbr.mPath, "email"); err != nil {
		mbr.Email = ""
	}
	return mbr, nil
}

func (gj *Gamejam) SaveTeam(tm *Team) error {
	var err error
	if err = gj.m.openDB(); err != nil {
		return err
	}
	defer gj.m.closeDB()

	// Save team data
	if err = gj.m.bolt.SetValue(tm.mPath, "name", tm.Name); err != nil {
		return err
	}

	// Save team members
	for _, mbr := range tm.Members {
		if err = gj.m.bolt.SetValue(mbr.mPath, "name", mbr.Name); err != nil {
			return err
		}
		if err = gj.m.bolt.SetValue(mbr.mPath, "slackid", mbr.SlackId); err != nil {
			return err
		}
		if err = gj.m.bolt.SetValue(mbr.mPath, "twitter", mbr.Twitter); err != nil {
			return err
		}
		if err = gj.m.bolt.SetValue(mbr.mPath, "email", mbr.Email); err != nil {
			return err
		}
	}

	// Save team game
	return gj.SaveGame(tm.Game)
}

// Delete the team tm
// TODO: Deletes should be done all at once when syncing memory to the DB
/*
func (gj *Gamejam) DeleteTeam(tm *Team) error {
	var err error
	if err = gj.m.openDB(); err != nil {
		return err
	}
	defer gj.m.closeDB()

	if len(tm.mPath) < 2 {
		return errors.New("Invalid team path: " + string(tm.mPath))
	}
	return gj.m.bolt.DeleteBucket(tm.mPath[:len(tm.mPath)-1], tm.UUID)
}
*/

// Delete the TeamMember mbr from Team tm
// TODO: Deletes should be done all at once when syncing memory to the DB
/*
func (gj *Gamejam) DeleteTeamMember(tm *Team, mbr *TeamMember) error {
	var err error
	if err = gj.m.openDB(); err != nil {
		return err
	}
	defer gj.m.closeDB()

	if len(mbr.mPath) < 2 {
		return errors.New("Invalid team path: " + string(tm.mPath))
	}
	return gj.m.bolt.DeleteBucket(mbr.mPath[:len(mbr.mPath)-1], mbr.UUID)
}
*/

/**
 * In Memory functions
 * This is generally how the app accesses data
 */

// Add a team
func (gj *Gamejam) AddTeam(tm *Team) error {
	if _, err := gj.GetTeamById(tm.UUID); err == nil {
		return errors.New("A team with that ID already exists")
	}
	if _, err := gj.GetTeamByName(tm.Name); err == nil {
		return errors.New("A team with that Name already exists")
	}
	gj.Teams = append(gj.Teams, *tm)
	return nil
}

// Find a team by name
func (gj *Gamejam) GetTeamByName(nm string) (*Team, error) {
	for i := range gj.Teams {
		if gj.Teams[i].Name == nm {
			return &gj.Teams[i], nil
		}
	}
	return nil, errors.New("Invalid team name given")
}

// Remove a team by id
func (gj *Gamejam) RemoveTeamById(id string) error {
	idx := -1
	for i := range gj.Teams {
		if gj.Teams[i].UUID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return errors.New("Invalid Team ID given")
	}
	gj.Teams = append(gj.Teams[:idx], gj.Teams[idx+1:]...)
	return nil
}
