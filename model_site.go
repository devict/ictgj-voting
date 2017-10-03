package main

const (
	SiteModeWaiting = iota
	SiteModeVoting
	SiteModeError
)

// SiteData is stuff that stays the same
type siteData struct {
	Title       string
	Port        int
	SessionName string
	ServerDir   string
	DevMode     bool
	Mode        int

	CurrentJam string

	Teams []Team
	Votes []Vote
}

// NewSiteData returns a siteData object with the default values
func NewSiteData() *siteData {
	ret := new(siteData)
	ret.Title = "ICT GameJam"
	ret.Port = 8080
	ret.SessionName = "ict-gamejam"
	ret.ServerDir = "./"
	return ret
}

func (s *siteData) getTeamByUUID(uuid string) *Team {
	for i := range s.Teams {
		if s.Teams[i].UUID == uuid {
			return &s.Teams[i]
		}
	}
	return nil
}

// save 's' to the database
func (s *siteData) save() error {
	var err error
	if err = db.open(); err != nil {
		return err
	}
	defer db.close()

	siteConf := []string{"site"}
	if err = db.bolt.SetValue(siteConf, "title", s.Title); err != nil {
		return err
	}
	if err = db.bolt.SetInt(siteConf, "port", s.Port); err != nil {
		return err
	}
	if err = db.bolt.SetValue(siteConf, "session-name", s.SessionName); err != nil {
		return err
	}
	return db.bolt.SetValue(siteConf, "server-dir", s.ServerDir)
}
