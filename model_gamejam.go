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

	m       *model
	updates []string
}

func NewGamejam(m *model) *Gamejam {
	gj := new(Gamejam)
	gj.m = m
	return gj
}

func (m *model) LoadCurrentJam() *Gamejam {
	if err := m.openDB(); err != nil {
		return err
	}
	defer m.closeDB()

	var err error
	jamPath := []string{"jam"}
	gj := NewGamejam(m)
	gj.Name, _ = m.bolt.GetValue(jamPath, "name")

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
	if err := s.m.openDB(); err != nil {
		return err
	}
	defer s.m.closeDB()

	for i := range updates {
		// TODO: Save
	}
}
