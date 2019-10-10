package main

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/br0xen/boltease"
	"github.com/pborman/uuid"
)

// Archived GameJams are in their own DB files in the data directory
// named `gamejam_<uuid>.db`
type Archive struct {
	Jams []ArchivedGamejam

	m     *model   // The model that holds the main archive data
	mPath []string // The path in the (main) db to the archives
}

func NewArchive(m *model) *Archive {
	arc := new(Archive)
	arc.m = m
	arc.mPath = []string{"archive"}
	return arc
}

func (m *model) LoadArchive() (*Archive, error) {
	if err := m.openDB(); err != nil {
		return nil, err
	}
	defer m.closeDB()

	arc := NewArchive(m)
	vals, err := m.bolt.GetValueList(arc.mPath)
	if err != nil {
		// There apparently aren't any archived Jams
		return arc, nil
	}
	for _, v := range vals {
		arcgj, err := NewArchivedGamejam(v)
		if err == nil {
			arc.Jams = append(arc.Jams, *arcgj)
		}
	}

	return arc, nil
}

func (m *model) SaveArchive() error {
	var err error
	if err = m.openDB(); err != nil {
		return err
	}
	defer m.closeDB()
	fmt.Println("Saving Archive")
	for k, v := range m.archive.Jams {
		fmt.Printf("> %d. %s\n", k, v.UUID)
		m.bolt.SetValue([]string{"archive"}, strconv.Itoa(k), v.UUID)
	}
	return nil
}

func (m *model) ArchiveCurrentJam() error {
	gj := new(ArchivedGamejam)
	if m.jam.UUID == "" {
		m.jam.UUID = uuid.New()
	}
	gj.UUID = m.jam.UUID
	gj.Name = m.jam.Name
	// We save the teams to the archive in their ranked order
	for k := range m.jam.Teams {
		gj.Teams = append(gj.Teams, m.jam.Teams[k])
	}
	for k := range m.jam.Votes {
		gj.Votes = append(gj.Votes, m.jam.Votes[k])
	}
	rankings := getCondorcetResult()
	for _, v := range rankings {
		for _, tv := range v.Teams {
			gj.Rankings = append(gj.Rankings, tv.UUID)
		}
	}

	err := gj.Save()
	if err != nil {
		return err
	}
	m.archive.Jams = append(m.archive.Jams, *gj)
	// Now we need to clear the current jam
	m.jam = NewGamejam(m)

	// Delete the Teams/Votes buckets from the jam
	if err := m.openDB(); err != nil {
		return err
	}
	defer m.closeDB()
	if err := m.bolt.DeleteBucket([]string{"jam"}, "teams"); err != nil {
		return err
	}
	if err := m.bolt.DeleteBucket([]string{"jam"}, "votes"); err != nil {
		return err
	}
	return m.saveChanges()
}

type ArchivedGamejam struct {
	UUID     string
	Name     string
	Date     time.Time
	Rankings []string
	Teams    []Team
	Votes    []Vote
}

func NewArchivedGamejam(uuid string) (*ArchivedGamejam, error) {
	gj := new(ArchivedGamejam)
	gj.UUID = uuid
	bolt, err := boltease.Create(DataDir+"/gamejam_"+uuid+".db", 0600, nil)
	defer bolt.CloseDB()
	if err != nil {
		return nil, err
	}
	gj.Name, err = bolt.GetValue([]string{"jam"}, "name")
	if err != nil {
		return nil, err
	}
	// Now load in all of the teams
	var tmUUIDs []string
	if tmUUIDs, err = bolt.GetBucketList([]string{"jam", "teams"}); err != nil {
		return nil, err
	}
	for _, v := range tmUUIDs {
		tm, err := gj.LoadTeam(bolt, v)
		if err != nil {
			return nil, err
		}
		gj.Teams = append(gj.Teams, *tm)
	}

  // And finally, the Rankings
  var ranks []string
  if ranks, err = bolt.GetKeyList([]string{"jam", "rankings"}); err != nil {
    return nil, err
  }
  for _, v := range ranks {
    var tmUUID string
    tmUUID, err = bolt.GetValue([]string{"jam", "rankings"}, v)
    if err != nil {
      return nil, err
    }
    gj.Rankings = append(gj.Rankings, tmUUID)
  }

  // We could pull votes too... But I'm not right now.
	return gj, nil
}

func (a *ArchivedGamejam) LoadTeam(openbolt *boltease.DB, uuid string) (*Team, error) {
	tm := NewTeam(uuid)
	var err error
	if tm.Name, err = openbolt.GetValue(tm.mPath, "name"); err != nil {
		return nil, errors.New("Error loading team: " + err.Error())
	}

	// Load the Team Members
	var memberUuids []string
	mbrsPath := append(tm.mPath, "members")
	if memberUuids, err = openbolt.GetBucketList(mbrsPath); err == nil {
		for _, v := range memberUuids {
			mbr, err := NewTeamMember(uuid, v)
			if err != nil {
				return nil, err
			}
			if mbr.Name, err = openbolt.GetValue(mbr.mPath, "name"); err != nil {
				return nil, errors.New("Error loading team member: " + err.Error())
			}
			tm.Members = append(tm.Members, *mbr)
		}
	}

	// Load the Team's Game
	tm.Game, err = NewGame(uuid)
	if err != nil {
		return nil, err
	}
	if tm.Game.Name, err = openbolt.GetValue(tm.Game.mPath, "name"); err != nil {
		tm.Game.Name = ""
	}
	if tm.Game.Link, err = openbolt.GetValue(tm.Game.mPath, "link"); err != nil {
		tm.Game.Link = ""
	}
	return tm, nil
}

func (a *ArchivedGamejam) Save() error {
	bolt, err := boltease.Create(DataDir+"/gamejam_"+a.UUID+".db", 0600, nil)
	defer bolt.CloseDB()
	if err != nil {
		return err
	}
	// Gamejam info
	if err := bolt.SetValue([]string{"jam"}, "uuid", a.UUID); err != nil {
		return err
	}
	if err := bolt.SetValue([]string{"jam"}, "name", a.Name); err != nil {
		return err
	}
	// Teams info
	for _, tm := range a.Teams {
		if err := bolt.SetValue(tm.mPath, "name", tm.Name); err != nil {
			return err
		}
		for _, mbr := range tm.Members {
			if err = bolt.SetValue(mbr.mPath, "name", mbr.Name); err != nil {
				return err
			}
			if err = bolt.SetValue(mbr.mPath, "slackid", mbr.SlackId); err != nil {
				return err
			}
			if err = bolt.SetValue(mbr.mPath, "twitter", mbr.Twitter); err != nil {
				return err
			}
			if err = bolt.SetValue(mbr.mPath, "email", mbr.Email); err != nil {
				return err
			}
		}
		// The team's game
		gm := tm.Game
		if err := bolt.MkBucketPath(gm.mPath); err != nil {
			return err
		}

		if err := bolt.SetValue(gm.mPath, "name", gm.Name); err != nil {
			return err
		}
		if err := bolt.SetValue(gm.mPath, "link", gm.Link); err != nil {
			return err
		}
		if err := bolt.SetValue(gm.mPath, "description", gm.Description); err != nil {
			return err
		}
		if err := bolt.SetValue(gm.mPath, "framework", gm.Framework); err != nil {
			return err
		}
		// Save screenshots
		if err := bolt.MkBucketPath(append(gm.mPath, "screenshots")); err != nil {
			return err
		}

		for _, ss := range gm.Screenshots {
			if err = bolt.MkBucketPath(ss.mPath); err != nil {
				return err
			}
			if err = bolt.SetValue(ss.mPath, "description", ss.Description); err != nil {
				return err
			}
			if err = bolt.SetValue(ss.mPath, "image", ss.Image); err != nil {
				return err
			}
			if err = bolt.SetValue(ss.mPath, "filetype", ss.Filetype); err != nil {
				return err
			}
		}
	}
	// All teams are archived
	// Move on to votes
	for _, vt := range a.Votes {
		for _, v := range vt.Choices {
			bolt.SetValue(vt.mPath, strconv.Itoa(v.Rank), v.Team)
		}
		bolt.SetValue(vt.mPath, "voterstatus", vt.VoterStatus)
		bolt.SetValue(vt.mPath, "discovery", vt.Discovery)
	}
	// And the rankings
	if err = bolt.MkBucketPath([]string{"jam", "rankings"}); err != nil {
		return err
	}
	for kr, vr := range a.Rankings {
		if err = bolt.SetValue([]string{"jam", "rankings"}, strconv.Itoa(kr), vr); err != nil {
			return err
		}
	}

	return nil
}
