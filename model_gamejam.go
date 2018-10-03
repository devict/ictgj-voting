package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

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

	m     *model   // The model that holds this gamejam's data
	mPath []string // The path in the db to this gamejam

	IsChanged bool // Flag to tell if we need to update the db
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

func (m *model) LoadCurrentJam() (*Gamejam, error) {
	if err := m.openDB(); err != nil {
		return nil, err
	}
	defer m.closeDB()

	gj := NewGamejam(m)
	gj.Name, _ = m.bolt.GetValue(gj.mPath, "name")

	// Load all teams
	gj.Teams = gj.LoadAllTeams()

	// Load all votes
	gj.Votes = gj.LoadAllVotes()

	return gj, nil
}

// Save everything to the DB whether it's flagged as changed or not
func (gj *Gamejam) SaveToDB() error {
	if err := gj.m.openDB(); err != nil {
		return err
	}
	defer gj.m.closeDB()

	var errs []error
	if err := gj.m.bolt.SetValue(gj.mPath, "name", gj.Name); err != nil {
		errs = append(errs, err)
	}
	// Save all Teams
	for _, tm := range gj.Teams {
		fmt.Println("Saving Team " + tm.Name + " data to DB")
		if err := gj.SaveTeam(&tm); err != nil {
			errs = append(errs, err)
		}
	}

	// Save all Votes
	for _, vt := range gj.Votes {
		if err := gj.SaveVote(&vt); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		var errTxt string
		for i := range errs {
			errTxt = errTxt + errs[i].Error() + "\n"
		}
		errTxt = strings.TrimSpace(errTxt)
		return errors.New("Error(s) saving to DB: " + errTxt)
	}
	return nil
}
