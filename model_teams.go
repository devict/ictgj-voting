package main

import "errors"

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
	return &Team{
		UUID:  id,
		mPath: []string{"jam", "teams", id},
	}
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
func NewTeamMember(tmId, uId string) *TeamMember {
	return &TeamMember{
		UUID:  uId,
		mPath: []string{"jam", "teams", tmId, "members", uId},
	}
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
		return err
	}
	defer gj.m.closeDB()

	if tmUUIDs, err = m.bolt.GetBucketList(mbrsPath); err != nil {
		return ret
	}
	for _, v := range tmUUIDs {
		tm := gj.LoadTeam(v)
		if tm != nil {
			ret = append(ret, tm)
		}
	}
	return ret
}

// Load a team out of the database
func (gj *Gamejam) LoadTeam(uuid string) *Team {
	var err error
	if err = gj.m.openDB(); err != nil {
		return err
	}
	defer gj.m.closeDB()

	// Team Data
	tm := NewTeam(uuid)
	if tm.Name, err = gj.m.bolt.GetValue(tm.mPath, "name"); err != nil {
		return nil
	}

	// Team Members
	tm.Members = gj.LoadTeamMembers(uuid)

	// Team Game
	tm.Game = gj.LoadTeamGame(uuid)

	return tm
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
			mbr := gj.LoadTeamMember(tmId, v)
			if mbr != nil {
				ret = append(ret, mbr)
			}
		}
	}
	return ret
}

// Load a team member from the DB and return it
func (gj *Gamejam) LoadTeamMember(tmId, mbrId string) *TeamMember {
	var err error
	if err = gj.m.openDB(); err != nil {
		return nil
	}
	defer gj.m.closeDB()

	mbr := NewTeamMember(tmId, mbrId)
	// Name is the only required field
	if mbr.Name, err = gj.m.bolt.GetValue(mbr.mPath, "name"); err != nil {
		return nil
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
	return mbr
}

func (gj *Gamejam) SaveTeam(tm *Team) error {
	var err error
	if err = gj.m.openDB(); err != nil {
		return err
	}
	defer gj.m.closeDB()

	// Save team data
	if err = gj.m.bolt.SetValue(tm.mPath, "name"); err != nil {
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
	return gj.SaveGame(gm)
}

// Delete the team tm
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

// Delete the TeamMember mbr from Team tm
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

/**
 * In Memory functions
 * This is generally how the app accesses data
 */

// Find a team by it's ID
func (gj *Gamejam) GetTeamById(id string) *Team {
	for i := range gj.Teams {
		if gj.Teams[i].UUID == id {
			return gj.Teams[i]
		}
	}
	return nil
}

// Find a team by name
func (gj *Gamejam) GetTeamByName(nm string) *Team {
	for i := range gj.Teams {
		if gj.Teams[i].Name == nm {
			return gj.Teams[i]
		}
	}
	return nil
}
