package main

import (
	"errors"

	"github.com/pborman/uuid"
)

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

	mPath []string // The path in the DB to this game
}

// Create a new game object
func NewGame(tmId string) (*Game, error) {
	if tmId == "" {
		return nil, errors.New("Team ID is required")
	}
	return &Game{
		TeamId: tmId,
		mPath:  []string{"jam", "teams", tmId, "game"},
	}, nil
}

func (gm *Game) GetScreenshot(ssId string) (*Screenshot, error) {
	for _, ss := range gm.Screenshots {
		if ss.UUID == ssId {
			return &ss, nil
		}
	}
	return nil, errors.New("Invalid Id")
}

func (gm *Game) RemoveScreenshot(ssId string) error {
	idx := -1
	for i, ss := range gm.Screenshots {
		if ss.UUID == ssId {
			idx = i
			return nil
		}
	}
	if idx < 0 {
		return errors.New("Invalid Id")
	}
	gm.Screenshots = append(gm.Screenshots[:idx], gm.Screenshots[idx+1:]...)
	return nil
}

type Screenshot struct {
	UUID        string
	Description string
	Image       string
	Thumbnail   string
	Filetype    string

	mPath []string // The path in the DB to this screenshot
}

// Create a Screenshot Object
func NewScreenshot(tmId, ssId string) (*Screenshot, error) {
	if tmId == "" {
		return nil, errors.New("Team ID is required")
	}
	if ssId == "" {
		// Generate a new UUID
		ssId = uuid.New()
	}
	return &Screenshot{
		UUID:  ssId,
		mPath: []string{"jam", "teams", tmId, "game", "screenshots", ssId},
	}, nil
}

/**
 * DB Functions
 * These are generally just called when the app starts up, or when the periodic 'save' runs
 */

// Load a team's game from the DB and return it
func (gj *Gamejam) LoadTeamGame(tmId string) (*Game, error) {
	var err error
	if err = gj.m.openDB(); err != nil {
		return nil, err
	}
	defer gj.m.closeDB()

	gm, err := NewGame(tmId)
	if err != nil {
		return nil, err
	}
	if gm.Name, err = gj.m.bolt.GetValue(gm.mPath, "name"); err != nil {
		gm.Name = ""
	}
	if gm.Description, err = gj.m.bolt.GetValue(gm.mPath, "description"); err != nil {
		gm.Description = ""
	}
	if gm.Link, err = gj.m.bolt.GetValue(gm.mPath, "link"); err != nil {
		gm.Link = ""
	}

	// Now get the game screenshots
	gm.Screenshots = gj.LoadTeamGameScreenshots(tmId)

	return gm, nil
}

// Load a games screenshots from the DB
func (gj *Gamejam) LoadTeamGameScreenshots(tmId string) []Screenshot {
	var err error
	if err = gj.m.openDB(); err != nil {
		return nil
	}
	defer gj.m.closeDB()

	var ret []Screenshot
	gm, err := NewGame(tmId)
	if err != nil {
		return ret
	}
	ssBktPath := append(gm.mPath, "screenshots")
	var ssIds []string
	ssIds, _ = gj.m.bolt.GetBucketList(ssBktPath)
	for _, v := range ssIds {
		ssLd, _ := gj.LoadTeamGameScreenshot(tmId, v)
		if ssLd != nil {
			ret = append(ret, *ssLd)
		}
	}
	return ret
}

// Load a screenshot from the DB
func (gj *Gamejam) LoadTeamGameScreenshot(tmId, ssId string) (*Screenshot, error) {
	var err error
	if err = gj.m.openDB(); err != nil {
		return nil, err
	}
	defer gj.m.closeDB()

	ret, err := NewScreenshot(tmId, ssId)
	if err != nil {
		return nil, err
	}
	if ret.Description, err = gj.m.bolt.GetValue(ret.mPath, "description"); err != nil {
		return nil, err
	}
	if ret.Image, err = gj.m.bolt.GetValue(ret.mPath, "image"); err != nil {
		return nil, err
	}
	if ret.Thumbnail, err = gj.m.bolt.GetValue(ret.mPath, "thumbnail"); err != nil {
		return nil, err
	}
	if ret.Thumbnail == "" {
		ret.Thumbnail = ret.Image
	}
	if ret.Filetype, err = gj.m.bolt.GetValue(ret.mPath, "filetype"); err != nil {
		return nil, err
	}
	return ret, err
}

// Save a game to the DB
func (gj *Gamejam) SaveGame(gm *Game) error {
	var err error
	if err = gj.m.openDB(); err != nil {
		return err
	}
	defer gj.m.closeDB()

	if err := gj.m.bolt.MkBucketPath(gm.mPath); err != nil {
		return err
	}

	var tm *Team
	if tm, err = gj.GetTeamById(gm.TeamId); err != nil {
		return err
	}
	if gm.Name == "" {
		gm.Name = tm.Name + "'s Game"
	}
	if err := gj.m.bolt.SetValue(gm.mPath, "name", gm.Name); err != nil {
		return err
	}
	if err := gj.m.bolt.SetValue(gm.mPath, "link", gm.Link); err != nil {
		return err
	}
	if err := gj.m.bolt.SetValue(gm.mPath, "description", gm.Description); err != nil {
		return err
	}
	if err := gj.m.bolt.MkBucketPath(append(gm.mPath, "screenshots")); err != nil {
		return err
	}
	return gj.SaveScreenshots(gm)
}

// Save all of the game's screenshots to the DB
// Remove screenshots from the DB that aren't in the game object
func (gj *Gamejam) SaveScreenshots(gm *Game) error {
	var err error
	if err = gj.m.openDB(); err != nil {
		return err
	}
	defer gj.m.closeDB()

	for _, ss := range gm.Screenshots {
		if err = gj.SaveScreenshot(&ss); err != nil {
			return err
		}
	}
	// Now remove unused screenshots
	ssPath := append(gm.mPath, "screenshots")
	var ssIds []string
	if ssIds, err = gj.m.bolt.GetBucketList(ssPath); err != nil {
		return err
	}
	for i := range ssIds {
		ss, _ := gm.GetScreenshot(ssIds[i])
		if ss != nil {
			// A valid screenshot, next
			continue
		}
		if ss, err = NewScreenshot(gm.TeamId, ssIds[i]); err != nil {
			// Error building screenshot to delete...
			continue
		}
		if err = gj.DeleteScreenshot(ss); err != nil {
			return err
		}
	}
	return nil
}

// Save a screenshot
func (gj *Gamejam) SaveScreenshot(ss *Screenshot) error {
	var err error
	if err = gj.m.openDB(); err != nil {
		return err
	}
	defer gj.m.closeDB()

	if err = gj.m.bolt.MkBucketPath(ss.mPath); err != nil {
		return err
	}
	if err = gj.m.bolt.SetValue(ss.mPath, "description", ss.Description); err != nil {
		return err
	}
	if err = gj.m.bolt.SetValue(ss.mPath, "image", ss.Image); err != nil {
		return err
	}
	if err = gj.m.bolt.SetValue(ss.mPath, "filetype", ss.Filetype); err != nil {
		return err
	}
	return nil
}

// Delete a screenshot
func (gj *Gamejam) DeleteScreenshot(ss *Screenshot) error {
	var err error
	if err = gj.m.openDB(); err != nil {
		return nil
	}
	defer gj.m.closeDB()

	ssPath := ss.mPath[:len(ss.mPath)-1]
	return gj.m.bolt.DeleteBucket(ssPath, ss.UUID)
}

/**
 * In Memory functions
 * This is generally how the app accesses client data
 */

// Set the given team's game to gm
func (gj *Gamejam) UpdateGame(tmId string, gm *Game) error {
	tm, err := gj.GetTeamById(tmId)
	if err != nil {
		return errors.New("Error getting team: " + err.Error())
	}
	tm.Game = gm
	gj.IsChanged = true
	return nil
}
