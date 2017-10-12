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
	updates []string
}

func NewGamejam(m *model) *Gamejam {
	gj := new(Gamejam)
	gj.m = m
	gj.mPath = []string{"jam"}
	return gj
}

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

func (gj *Gamejam) getTeamByUUID(uuid string) *Team {
	for i := range gj.Teams {
		if gj.Teams[i].UUID == uuid {
			return &gj.Teams[i]
		}
	}
	return nil
}

func (gj *Gamejam) needsSave() bool {
	return len(updates) > 0
}

func (gj *Gamejam) saveToDB() error {
	if err := gj.m.openDB(); err != nil {
		return err
	}
	defer gj.m.closeDB()

	for i := range updates {
		// TODO: Save
	}
}
