package main

import (
	"errors"

	"github.com/br0xen/boltease"
)

// model stores the current jam in memory, and has the ability to access archived dbs
type model struct {
	bolt       *boltease.DB
	dbOpened   int
	dbFileName string

	site    *siteData // Configuration data for the site
	jam     *Gamejam  // The currently active gamejam
	clients []Client  // Web clients that have connected to the server
}

// Update Flags: Which parts of the model need to be updated
const (
	UpdateSiteData = iota
	UpdateJamData
)

func NewModel() (*model, error) {
	var err error
	m := new(model)

	m.dbFileName = DbName
	if err = m.openDB(); err != nil {
		return nil, errors.New("Unable to open DB: " + err.Error())
	}
	defer m.closeDB()

	// Initialize the DB
	if err = m.initDB(); err != nil {
		return nil, errors.New("Unable to initialize DB: " + err.Error())
	}

	// Load the site data
	m.site = m.LoadSiteData()

	// Load the jam data
	m.jam = m.LoadCurrentJam()

	// Load web clients
	m.clients = m.LoadAllClients()

	return &m, nil
}

func (m *model) openDB() error {
	m.dbOpened += 1
	if db.dbOpened == 1 {
		var err error
		m.bolt, err = boltease.Create(m.dbFileName, 0600, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *model) closeDB() error {
	m.dbOpened -= 1
	if m.dbOpened == 0 {
		return m.bolt.CloseDB()
	}
	return nil
}

func (m *model) initDB() error {
	var err error
	if err = m.openDB(); err != nil {
		return err
	}
	defer m.closeDB()

	// Create the path to the bucket to store admin users
	if err = m.bolt.MkBucketPath([]string{"users"}); err != nil {
		return err
	}
	// Create the path to the bucket to store the web clients
	if err = m.bolt.MkBucketPath([]string{"clients"}); err != nil {
		return err
	}
	// Create the path to the bucket to store the current jam & teams
	if err = m.bolt.MkBucketPath([]string{"jam", "teams"}); err != nil {
		return err
	}
	// Create the path to the bucket to store the list of archived jams
	if err = m.bolt.MkBucketPath([]string{"archive"}); err != nil {
		return err
	}
	// Create the path to the bucket to store site config data
	return m.bolt.MkBucketPath([]string{"site"})
}

// saveChanges saves any parts of the model that have been flagged as changed to the database
func (m *model) saveChanges() error {
	var err error
	if err = m.openDB(); err != nil {
		return err
	}
	defer m.closeDB()

	if m.site.needsSave() {
		if err = m.site.saveToDB(); err != nil {
			return err
		}
	}
	if m.jam.needsSave() {
		if err = m.jam.saveToDB(); err != nil {
			return err
		}
	}
	if m.clientsUpdated {
		if err = m.SaveAllClients(); err != nil {
			return err
		}
	}
	return nil
}
