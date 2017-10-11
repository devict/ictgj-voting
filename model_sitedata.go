package main

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
	changed bool
}

// NewSiteData returns a siteData object with the default values
func NewSiteData(m *model) *siteData {
	ret := new(siteData)
	ret.Title = "ICT GameJam"
	ret.Port = 8080
	ret.SessionName = "ict-gamejam"
	ret.ServerDir = "./"
	ret.m = m
	return ret
}

// Mode flags for how the site is currently running
const (
	SiteModeWaiting = iota
	SiteModeVoting
	SiteModeError
)

// load the site data out of the database
// If fields don't exist in the DB, don't clobber what is already in s
func (s *siteData) loadFromDB() error {
	if err := s.m.openDB(); err != nil {
		return err
	}
	defer s.m.closeDB()

	siteConf := []string{"site"}
	if title, err := s.m.bolt.GetValue(siteConf, "title"); err == nil {
		s.Title = title
	}
	if port, err := s.m.bolt.GetInt(siteConf, "port"); err == nil {
		s.Port = port
	}
	if sessionName, err = s.m.bolt.GetValue(siteConf, "session-name"); err == nil {
		s.SessionName = sessionName
	}
	if serverDir, err = s.m.bolt.GetValue(siteConf, "server-dir"); err == nil {
		s.ServerDir = serverDir
	}
	s.changed = false
	return nil
}

func (s *siteData) needsSave() bool {
	return s.changed
}

func (s *siteData) saveToDB() error {
	if err := s.m.openDB(); err != nil {
		return err
	}
	defer s.m.closeDB()

	siteConf := []string{"site"}
	if err = s.m.bolt.SetValue(siteConf, "title", s.Title); err != nil {
		return err
	}
	if err = s.m.bolt.SetInt(siteConf, "port", s.Port); err != nil {
		return err
	}
	if err = s.m.bolt.SetValue(siteConf, "session-name", s.SessionName); err != nil {
		return err
	}
	if err = s.m.bolt.SetValue(siteConf, "server-dir", s.ServerDir); err != nil {
		return err
	}
	s.changed = false
	return nil
}

func (s *siteData) getAuthMode() int {
	return s.authMode
}

func (s *siteData) setAuthMode(mode int) {
	if mode != s.authMode {
		s.authMode = mode
		s.changed = true
	}
}

func (s *siteData) getPublicMode() int {
	return s.publicMode
}

func (s *siteData) setPublicMode(mode int) {
	if mode != s.publicMode {
		s.publicMode = mode
		s.changed = true
	}
}
