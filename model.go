package main

import (
	"errors"
	"fmt"
	"os"

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
	archive *Archive  // The archive of past game jams

	clientsUpdated bool
}

// Update Flags: Which parts of the model need to be updated
const (
	UpdateSiteData = iota
	UpdateJamData
)

func NewModel() (*model, error) {
	var err error
	m := new(model)

	// make sure the data directory exists
	if err = os.MkdirAll(DataDir, os.ModePerm); err != nil {
		return nil, errors.New("Unable to create Data Directory: " + err.Error())
	}
	m.dbFileName = DataDir + "/" + DbName
	if err = m.openDB(); err != nil {
		return nil, errors.New("Unable to open DB: " + err.Error())
	}
	defer m.closeDB()

	// Initialize the DB
	if err = m.initDB(); err != nil {
		return nil, errors.New("Unable to initialize DB: " + err.Error())
	}

	// Load the site data
	m.site = NewSiteData(m)
	if err = m.site.LoadFromDB(); err != nil {
		// Error loading from the DB, set to defaults
		def := NewSiteData(m)
		m.site = def
	}

	// Load the jam data
	if m.jam, err = m.LoadCurrentJam(); err != nil {
		return nil, errors.New("Unable to load current jam: " + err.Error())
	}

	// Load web clients
	m.clients = m.LoadAllClients()

	// Load the archives
	if m.archive, err = m.LoadArchive(); err != nil {
		return nil, errors.New("Unable to load game jam archive: " + err.Error())
	}

	return m, nil
}

func (m *model) openDB() error {
	m.dbOpened += 1
	if m.dbOpened == 1 {
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

	fmt.Println("Saving Site data to DB")
	if err = m.site.SaveToDB(); err != nil {
		return err
	}
	fmt.Println("Saving Jam data to DB")
	if err = m.jam.SaveToDB(); err != nil {
		return err
	}
	m.jam.IsChanged = false
	fmt.Println("Saving Client data to DB")
	if err = m.SaveAllClients(); err != nil {
		return err
	}
	m.clientsUpdated = false
	if err = m.SaveArchive(); err != nil {
		return err
	}
	return nil
}
