package main

import "errors"

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
func NewGame(tmId string) *Game {
	return &Game{
		TeamId: tmId,
		mPath:  []string{"jam", "teams", tmId, "game"},
	}
}

func (gm *Game) GetScreenshot(ssId string) *Screenshot {
	for _, ss := range gm.Screenshots {
		if ss.UUID == ssId {
			return ss
		}
	}
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
func NewScreenshot(tmId, ssId string) *Screenshot {
	return &Screenshot{
		UUID:  ssId,
		mPath: []string{"jam", "teams", tmId, "game", "screenshots", ssId},
	}
}

/**
 * DB Functions
 * These are generally just called when the app starts up, or when the periodic 'save' runs
 */

// Load a team's game from the DB and return it
func (gj *Gamejam) LoadTeamGame(tmId string) *Game {
	var err error
	if err = gj.m.openDB(); err != nil {
		return nil
	}
	defer gj.m.closeDB()

	gm := NewGame(tmId)
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

	return &gm
}

// Load a games screenshots from the DB
func (gj *Gamejam) LoadTeamGameScreenshots(tmId string) []Screenshot {
	var err error
	if err = gj.m.openDB(); err != nil {
		return nil
	}
	defer gj.m.closeDB()

	var ret []Screenshot
	gm := NewGame(tmId)
	ssBktPath := append(gm.mPath, "screenshots")
	var ssIds []string
	ssIds, _ = gj.m.bolt.GetBucketList(ssBktPath)
	for _, v := range ssIds {
		ssLd := gj.LoadTeamGameScreenshot(tmId, v)
		if ssLd != nil {
			ret = append(ret, ssLd)
		}
	}
	return ret
}

// Load a screenshot from the DB
func (gj *Gamejam) LoadTeamGameScreenshot(tmId, ssId string) *Screenshot {
	var err error
	if err = gj.m.openDB(); err != nil {
		return nil
	}
	defer gj.m.closeDB()

	ret := NewScreenshot(tmId, ssId)
	if ret.Description, err = gj.m.bolt.GetValue(ret.mPath, "description"); err != nil {
		return nil
	}
	if ret.Image, err = gj.m.bolt.GetValue(ret.mPath, "image"); err != nil {
		return nil
	}
	if ret.Thumbnail, err = gj.m.bolt.GetValue(ret.mPath, "thumbnail"); err != nil {
		return nil
	}
	if ret.Thumbnail == "" {
		ret.Thumbnail = ret.Image
	}
	if ret.Filetype, err = gj.m.bolt.GetValue(ret.mPath, "filetype"); err != nil {
		return nil
	}
	return ret
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
		if err = gj.SaveScreenshot(gm.TeamId, ss); err != nil {
			return err
		}
	}
	// Now remove unused screenshots
	ssPath := append(gm.mPath, "screenshots")
	if ssIds, err = gj.m.bolt.GetBucketList(ssPath); err != nil {
		return err
	}
	for i := range ssIds {
		if gm.GetScreenshot(ssIds[i]) == nil {
			if err = gj.DeleteScreenshot(NewScreenshot(tm.TeamId, ssIds[i])); err != nil {
				return err
			}
		}
	}
}

// Save a screenshot
func (gj *Gamejam) SaveScreenshot(tmId string, ss *Screenshot) error {
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
	var found bool
	tm := gj.GetTeamById(tmId)
	if tm == nil {
		return errors.New("Invalid team ID: " + gm.TeamId)
	}
	tm.Game = gm
	gj.NeedsUpdate([]string{"team", tmId, "game"})
	return nil
}
