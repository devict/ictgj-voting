package main

import (
	"time"

	"github.com/br0xen/boltease"
)

type Gamejam struct {
	UUID  string
	Name  string
	Date  time.Time
	Teams []Team
	Votes []Vote

	db       *boltease.DB
	dbOpened int
}

// Archived Gamejam data is stored in it's own file to keep things nice and organized
func (gj *Gamejam) openDB() error {
	gj.dbOpened += 1
	if gj.dbOpened == 1 {
		var err error
		gj.db, err = boltease.Create(gj.UUID+".db", 0600, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (gj *Gamejam) closeDB() error {
	gj.dbOpened -= 1
	if gj.dbOpened == 0 {
		return gj.db.CloseDB()
	}
	return nil
}

// archiveGameJam creates a separate gamejam file and populates it with the
// given name, teams, and votes
func archiveGamejam(nm string, teams []Team, votes []Vote) error {
	// TODO
	return nil
}

// dbGetGamejam returns a gamejam with the given uuid
// or nil if it couldn't be found
func dbGetGamejam(id string) *Gamejam {
	var err error
	if err = openDatabase(); err != nil {
		return nil
	}
	defer closeDatabase()

	ret := Gamejam{UUID: id}
	// TODO: Load gamejam teams, other details
	return ret
}

// dbGetGamejamByName looks for a gamejam with the given name
// and returns it, or it returns nil if it couldn't find it
func dbGetGamejamByName(nm string) *Gamejam {
	var err error
	if err = openDatabase(); err != nil {
		return err
	}
	defer closeDatabase()

	var gjid string
	if gjs, err = db.GetBucketList([]string{"gamejams"}); err == nil {
		for _, v := range gjUids {
			tstNm, _ := db.GetValue([]string{"gamejams", v}, "name")
			if tstNm == nm {
				// We've got it
				gjid = v
				break
			}
		}
	}
	if gjid == "" {
		return nil
	}
	return dbGetGamejam(gjid)
}
