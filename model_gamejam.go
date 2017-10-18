package main

import "time"

/**
 * Gamejam
 * Gamejam is the struct for any gamejam (current or archived)
 */
type Gamejam struct {
	UUID  string
	Name  string
	Date  time.Time
	Teams []Team
	Votes []Vote

	m       *model   // The model that holds this gamejam's data
	mPath   []string // The path in the db to this gamejam
	changed bool     // Flag to tell if we need to update the db
}

func NewGamejam(m *model) *Gamejam {
	gj := new(Gamejam)
	gj.m = m
	gj.mPath = []string{"jam"}
	return gj
}

/**
 * DB Functions
 * These are generally just called when the app starts up, or when the periodic 'save' runs
 */

func (m *model) LoadCurrentJam() *Gamejam {
	if err := m.openDB(); err != nil {
		return err
	}
	defer m.closeDB()

	var err error
	gj := NewGamejam(m)
	gj.Name, _ = m.bolt.GetValue(gj.mPath, "name")

	// Load all teams
	gj.Teams = gj.LoadAllTeams()

	// Load all votes
	gj.Votes = gj.LoadAllVotes()

	return gj
}

// Save everything to the DB whether it's flagged as changed or not
func (gj *Gamejam) saveToDB() error {
	if err := gj.m.openDB(); err != nil {
		return err
	}
	defer gj.m.closeDB()

}

/**
 * In Memory functions
 * This is generally how the app accesses client data
 */
func (gj *Gamejam) getTeamByUUID(uuid string) *Team {
	for i := range gj.Teams {
		if gj.Teams[i].UUID == uuid {
			return &gj.Teams[i]
		}
	}
	return nil
}

// Check if pth is already in updates, if not, add it
func (gj *Gamejam) NeedsUpdate(pth []string) {
	var found bool
	for _, v := range gj.updates {
		if !(len(v) == len(pth)) {
			continue
		}
		// The lengths are the same, do all elements match?
		var nxt bool
		for i := range pth {
			if v[i] != pth[i] {
				nxt = true
			}
		}
		if !nxt {
			// This pth is already in the 'updates' list
			found = true
			break
		}
	}
	if !found {
		gj.updates = append(gj.updates, pth)
	}
}
