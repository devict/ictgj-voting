package main

import "strconv"

/**
 * SiteData
 * Contains configuration for the website
 */
type siteData struct {
	title       string
	port        int
	sessionName string
	serverDir   string
	authMode    int
	publicMode  int

	DevMode bool
	Mode    int

	m       *model
	mPath   []string // The path in the db to this site data
	changed bool
}

// NewSiteData returns a siteData object with the default values
func NewSiteData(m *model) *siteData {
	ret := new(siteData)
	ret.Title = "ICT GameJam"
	ret.Port = 8080
	ret.SessionName = "ict-gamejam"
	ret.ServerDir = "./"
	ret.mPath = []string{"site"}
	ret.m = m
	return ret
}

// Authentication Modes: Flags for which clients are able to vote
const (
	AuthModeAuthentication = iota
	AuthModeAll
	AuthModeError
)

// Mode flags for how the site is currently running
const (
	SiteModeWaiting = iota
	SiteModeVoting
	SiteModeError
)

// load the site data out of the database
// If fields don't exist in the DB, don't clobber what is already in s
func (s *siteData) LoadFromDB() error {
	if err := s.m.openDB(); err != nil {
		return err
	}
	defer s.m.closeDB()

	if title, err := s.m.bolt.GetValue(s.mPath, "title"); err == nil {
		s.Title = title
	}
	if port, err := s.m.bolt.GetInt(s.mPath, "port"); err == nil {
		s.Port = port
	}
	if sessionName, err = s.m.bolt.GetValue(s.mPath, "session-name"); err == nil {
		s.SessionName = sessionName
	}
	if serverDir, err = s.m.bolt.GetValue(s.mPath, "server-dir"); err == nil {
		s.ServerDir = serverDir
	}
	s.changed = false
	return nil
}

// Return if the site data in memory has changed
func (s *siteData) NeedsSave() bool {
	return s.changed
}

// Save the site data into the DB
func (s *siteData) SaveToDB() error {
	if err := s.m.openDB(); err != nil {
		return err
	}
	defer s.m.closeDB()

	if err = s.m.bolt.SetValue(s.mPath, "title", s.Title); err != nil {
		return err
	}
	if err = s.m.bolt.SetInt(s.mPath, "port", s.Port); err != nil {
		return err
	}
	if err = s.m.bolt.SetValue(s.mPath, "session-name", s.SessionName); err != nil {
		return err
	}
	if err = s.m.bolt.SetValue(s.mPath, "server-dir", s.ServerDir); err != nil {
		return err
	}
	s.changed = false
	return nil
}

// Return the Auth Mode
func (s *siteData) GetAuthMode() int {
	return s.authMode
}

// Set the auth mode
func (s *siteData) SetAuthMode(mode int) error {
	if mode < AuthModeAuthentication || mode >= AuthModeError {
		return errors.Error("Invalid Authentication Mode: " + strconv.Itoa(mode))
	}
	if mode != s.authMode {
		s.authMode = mode
		s.changed = true
	}
}

// Return the public site mode
func (s *siteData) GetPublicMode() int {
	return s.publicMode
}

// Set the public site mode
func (s *siteData) SetPublicMode(mode int) error {
	if mode < SiteModeWaiting || mode >= SiteModeError {
		return errors.Error("Invalid Public Mode: " + strconv.Itoa(mode))
	}
	if mode != s.publicMode {
		s.publicMode = mode
		s.changed = true
	}
}
