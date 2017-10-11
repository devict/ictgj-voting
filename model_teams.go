package main

import (
	"errors"

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
}

// Create a team
func NewTeam(nm string) *Team {
	return &Team{
		UUID: uuid.New(),
		Name: nm,
	}
}

type TeamMember struct {
	UUID    string
	Name    string
	SlackId string
	Twitter string
	Email   string
}

// Create a new team member, only a name is required
func NewTeamMember(nm string) *TeamMember {
	m := TeamMember{Name: nm}
	return &m
}

// LoadAllTeams loads all teams for the jam out of the database
func (gj *Gamejam) LoadAllTeams() []Team {
	var err error
	var ret []Team
	if err = gj.m.openDB(); err != nil {
		return err
	}
	defer gj.m.closeDB()

	teamsPath := []string{"jam", "teams"}
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
	tmPath := []string{"jam", "teams", uuid}
	tm := new(Team)
	tm.UUID = uuid
	if tm.Name, err = gj.m.bolt.GetValue(tmPath, "name"); err != nil {
		return nil
	}

	// Team Members
	tm.Members = gj.LoadTeamMembers(uuid)

	// Team Game
	tm.Game = gj.LoadTeamGame(uuid)

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
	mbrsPath := []string{"jam", "teams", tmId, "members"}
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

	mbr := new(TeamMember)
	mbr.UUID = v
	teamMbrPath := append(mbrsPath, mbr.UUID)
	// Name is the only required field
	if mbr.Name, err = gj.m.bolt.GetValue(teamMbrPath, "name"); err != nil {
		return nil
	}
	if mbr.SlackId, err = gj.m.bolt.GetValue(teamMbrPath, "slackid"); err != nil {
		mbr.SlackId = ""
	}
	if mbr.Twitter, err = gj.m.bolt.GetValue(teamMbrPath, "twitter"); err != nil {
		mbr.Twitter = ""
	}
	if mbr.Email, err = gj.m.bolt.GetValue(teamMbrPath, "email"); err != nil {
		mbr.Email = ""
	}
	return mbr
}

/**
 * OLD FUNCTIONS
 */

// NewTeam creates a team with name nm and stores it in the DB
func NewTeam(nm string) error {
	var err error
	if err = db.open(); err != nil {
		return err
	}
	defer db.close()

	// Generate a UUID
	uuid := uuid.New()
	teamPath := []string{"teams", uuid}

	if err := db.bolt.MkBucketPath(teamPath); err != nil {
		return err
	}
	if err := db.bolt.SetValue(teamPath, "name", nm); err != nil {
		return err
	}
	if err := db.bolt.MkBucketPath(append(teamPath, "members")); err != nil {
		return err
	}
	gamePath := append(teamPath, "game")
	if err := db.bolt.MkBucketPath(gamePath); err != nil {
		return err
	}
	if err := db.bolt.SetValue(append(gamePath), "name", ""); err != nil {
		return err
	}
	return db.bolt.MkBucketPath(append(gamePath, "screenshots"))
}

// getTeam returns a team with the given id, or nil
func (db *currJamDb) getTeam(id string) *Team {
	var err error
	if err = db.open(); err != nil {
		return nil
	}
	defer db.close()

	teamPath := []string{"teams", id}
	tm := new(Team)
	tm.UUID = id
	if tm.Name, err = db.bolt.GetValue(teamPath, "name"); err != nil {
		return nil
	}
	tm.Members = tm.getTeamMembers()
	tm.Game = tm.getGame()
	return tm
}

// This function returns the team for a specific member
func (db *currJamDb) getTeamForMember(mbrid string) (*Team, error) {
	var err error
	if err = db.open(); err != nil {
		return nil, err
	}
	defer db.close()

	teams := db.getAllTeams()
	for i := range teams {
		var tmMbrs []TeamMember
		tmMbrs = teams[i].getTeamMembers()
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

// getAllTeams returns all teams in the database
func (db *currJamDb) getAllTeams() []Team {
	var ret []Team
	var err error
	if err = db.open(); err != nil {
		return ret
	}
	defer db.close()

	teamPath := []string{"teams"}
	var teamUids []string
	if teamUids, err = db.bolt.GetBucketList(teamPath); err != nil {
		return ret
	}
	for _, v := range teamUids {
		if tm := db.getTeam(v); tm != nil {
			ret = append(ret, *tm)
		}
	}
	return ret
}

// getTeamByName returns a team with the given name or nil
func (db *currJamDb) getTeamByName(nm string) *Team {
	var err error
	if err = db.open(); err != nil {
		return nil
	}
	defer db.close()

	teamPath := []string{"teams"}
	var teamUids []string
	if teamUids, err = db.bolt.GetBucketList(teamPath); err != nil {
		for _, v := range teamUids {
			var name string
			if name, err = db.bolt.GetValue(append(teamPath, v), "name"); name == nm {
				return db.getTeam(v)
			}
		}
	}
	return nil
}

// save saves the team to the db
func (tm *Team) save() error {
	var err error
	if err = db.open(); err != nil {
		return err
	}
	defer db.close()

	teamPath := []string{"teams", tm.UUID}
	if err = db.bolt.SetValue(teamPath, "name", tm.Name); err != nil {
		return err
	}

	// TODO: Save Team Members
	// TODO: Save Team Game
	return nil
}

// delete removes the team from the database
func (tm *Team) delete() error {
	var err error
	if err = db.open(); err != nil {
		return err
	}
	defer db.close()

	teamPath := []string{"teams"}
	return db.bolt.DeleteBucket(teamPath, tm.UUID)
}

func (tm *Team) getGame() *Game {
	var err error
	if err = db.open(); err != nil {
		return nil
	}
	defer db.close()

	gamePath := []string{"teams", tm.UUID, "game"}
	gm := new(Game)
	if gm.Name, err = db.bolt.GetValue(gamePath, "name"); err != nil {
		gm.Name = ""
	}
	gm.TeamId = tm.UUID
	if gm.Description, err = db.bolt.GetValue(gamePath, "description"); err != nil {
		gm.Description = ""
	}
	if gm.Link, err = db.bolt.GetValue(gamePath, "link"); err != nil {
		gm.Link = ""
	}
	gm.Screenshots = tm.getScreenshots()
	return gm
}

// Screenshots are saved as base64 encoded pngs
func (tm *Team) saveScreenshot(ss *Screenshot) error {
	var err error
	if err = db.open(); err != nil {
		return nil
	}
	defer db.close()

	ssPath := []string{"teams", tm.UUID, "game", "screenshots"}
	// Generate a UUID for this screenshot
	uuid := uuid.New()
	ssPath = append(ssPath, uuid)
	if err := db.bolt.MkBucketPath(ssPath); err != nil {
		return err
	}
	if err := db.bolt.SetValue(ssPath, "description", ss.Description); err != nil {
		return err
	}
	if err := db.bolt.SetValue(ssPath, "image", ss.Image); err != nil {
		return err
	}
	if err := db.bolt.SetValue(ssPath, "thumbnail", ss.Thumbnail); err != nil {
		return err
	}
	if err := db.bolt.SetValue(ssPath, "filetype", ss.Filetype); err != nil {
		return err
	}
	return nil
}

func (tm *Team) deleteScreenshot(ssId string) error {
	var err error
	if err = db.open(); err != nil {
		return nil
	}
	defer db.close()

	ssPath := []string{"teams", tm.UUID, "game", "screenshots"}
	return db.bolt.DeleteBucket(ssPath, ssId)
}

func (tm *Team) getTeamMember(mbrId string) *TeamMember {
	var err error
	if err = db.open(); err != nil {
		return nil
	}
	defer db.close()

	mbr := new(TeamMember)
	mbr.UUID = mbrId
	teamMbrPath := []string{"teams", tm.UUID, "members", mbr.UUID}
	if mbr.Name, err = db.bolt.GetValue(teamMbrPath, "name"); err != nil {
		return nil
	}
	if mbr.SlackId, err = db.bolt.GetValue(teamMbrPath, "slackid"); err != nil {
		return nil
	}
	if mbr.Twitter, err = db.bolt.GetValue(teamMbrPath, "twitter"); err != nil {
		return nil
	}
	if mbr.Email, err = db.bolt.GetValue(teamMbrPath, "email"); err != nil {
		return nil
	}
	return mbr
}

func (tm *Team) getTeamMembers() []TeamMember {
	var ret []TeamMember
	var err error
	if err = db.open(); err != nil {
		return ret
	}
	defer db.close()

	teamPath := []string{"teams", tm.UUID, "members"}
	var memberUuids []string
	if memberUuids, err = db.bolt.GetBucketList(teamPath); err == nil {
		for _, v := range memberUuids {
			var mbr *TeamMember
			if mbr = tm.getTeamMember(v); mbr != nil {
				ret = append(ret, *mbr)
			}
		}
	}
	return ret
}

func (tm *Team) updateTeamMember(mbr *TeamMember) error {
	var err error
	if err = db.open(); err != nil {
		return err
	}
	defer db.close()

	if mbr.UUID == "" {
		mbrs := tm.getTeamMembers()
		if len(mbrs) > 0 {
			for i := range mbrs {
				if mbrs[i].Name == mbr.Name {
					mbr.UUID = mbrs[i].UUID
					break
				}
			}
		}
	}
	if mbr.UUID == "" {
		// It's really a new one
		mbr.UUID = uuid.New()
	}

	mbrPath := []string{"teams", tm.UUID, "members", mbr.UUID}
	if db.bolt.SetValue(mbrPath, "name", mbr.Name) != nil {
		return err
	}
	if db.bolt.SetValue(mbrPath, "slackid", mbr.SlackId) != nil {
		return err
	}
	if db.bolt.SetValue(mbrPath, "twitter", mbr.Twitter) != nil {
		return err
	}
	if db.bolt.SetValue(mbrPath, "email", mbr.Email) != nil {
		return err
	}
	return nil
}

// deleteTeamMember removes a member from the database
func (tm *Team) deleteTeamMember(mbr *TeamMember) error {
	var err error
	if err = db.open(); err != nil {
		return err
	}
	defer db.close()

	teamPath := []string{"teams", tm.UUID, "members"}
	return db.bolt.DeleteBucket(teamPath, mbr.UUID)
}
