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

// newTeam creates a team with name nm and stores it in the DB
func (db *gjDatabase) newTeam(nm string) error {
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
func (db *gjDatabase) getTeam(id string) *Team {
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
func (db *gjDatabase) getTeamForMember(mbrid string) (*Team, error) {
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
func (db *gjDatabase) getAllTeams() []Team {
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
func (db *gjDatabase) getTeamByName(nm string) *Team {
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

func (tm *Team) getScreenshots() []Screenshot {
	var ret []Screenshot
	var err error
	ssPath := []string{"teams", tm.UUID, "game", "screenshots"}
	var ssIds []string
	if ssIds, err = db.bolt.GetBucketList(ssPath); err != nil {
		return ret
	}
	for _, v := range ssIds {
		if ss := tm.getScreenshot(v); ss != nil {
			ret = append(ret, *ss)
		}
	}
	return ret
}

func (tm *Team) getScreenshot(ssId string) *Screenshot {
	var err error
	ssPath := []string{"teams", tm.UUID, "game", "screenshots", ssId}
	ret := new(Screenshot)
	ret.UUID = ssId
	if ret.Description, err = db.bolt.GetValue(ssPath, "description"); err != nil {
		return nil
	}
	if ret.Image, err = db.bolt.GetValue(ssPath, "image"); err != nil {
		return nil
	}
	if ret.Thumbnail, err = db.bolt.GetValue(ssPath, "thumbnail"); err != nil {
		return nil
	}
	if ret.Thumbnail == "" {
		ret.Thumbnail = ret.Image
	}
	if ret.Filetype, err = db.bolt.GetValue(ssPath, "filetype"); err != nil {
		return nil
	}
	return ret
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

type TeamMember struct {
	UUID    string
	Name    string
	SlackId string
	Twitter string
	Email   string
}

// Create a new team member, only a name is required
func newTeamMember(nm string) *TeamMember {
	m := TeamMember{Name: nm}
	return &m
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
