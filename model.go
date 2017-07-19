package main

import (
	"errors"
	"strings"

	"github.com/br0xen/boltease"
)

var db *boltease.DB
var dbOpened int

const (
	SiteModeWaiting = iota
	SiteModeVoting
	SiteModeError
)

const (
	AuthModeAuthentication = iota
	AuthModeAll
	AuthModeError
)

func GetDefaultSiteConfig() *siteData {
	ret := new(siteData)
	ret.Title = "ICT GameJam"
	ret.Port = 8080
	ret.SessionName = "ict-gamejam"
	ret.ServerDir = "./"
	return ret
}

func openDatabase() error {
	dbOpened += 1
	if dbOpened == 1 {
		var err error
		db, err = boltease.Create(DbName, 0600, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func closeDatabase() error {
	dbOpened -= 1
	if dbOpened == 0 {
		return db.CloseDB()
	}
	return nil
}

func initDatabase() error {
	var err error
	if err = openDatabase(); err != nil {
		return err
	}
	defer closeDatabase()

	// Create the path to the bucket to store admin users
	if err := db.MkBucketPath([]string{"users"}); err != nil {
		return err
	}
	// Create the path to the bucket to store jam informations
	if err := db.MkBucketPath([]string{"jam"}); err != nil {
		return err
	}
	// Create the path to the bucket to store site config data
	return db.MkBucketPath([]string{"site"})
}

func dbSetCurrentJam(name string) error {
	var err error
	if err = openDatabase(); err != nil {
		return err
	}
	defer closeDatabase()

	return db.SetValue([]string{"site"}, "current-jam", name)
}

func dbHasCurrentJam() bool {
	var err error
	if _, err = dbGetCurrentJam(); err != nil {
		return false
	}
	return true
}

func dbGetCurrentJam() (string, error) {
	var ret string
	var err error
	if err = openDatabase(); err != nil {
		return "", err
	}
	defer closeDatabase()

	ret, err = db.GetValue([]string{"site"}, "current-jam")

	if err == nil && strings.TrimSpace(ret) == "" {
		return ret, errors.New("No Jam Name Specified")
	}
	return ret, err
}

func dbGetSiteConfig() *siteData {
	var ret *siteData
	def := GetDefaultSiteConfig()
	var err error
	if err = openDatabase(); err != nil {
		return def
	}
	defer closeDatabase()

	ret = new(siteData)
	siteConf := []string{"site"}
	if ret.Title, err = db.GetValue(siteConf, "title"); err != nil {
		ret.Title = def.Title
	}
	if ret.Port, err = db.GetInt(siteConf, "port"); err != nil {
		ret.Port = def.Port
	}
	if ret.SessionName, err = db.GetValue(siteConf, "session-name"); err != nil {
		ret.SessionName = def.SessionName
	}
	if ret.ServerDir, err = db.GetValue(siteConf, "server-dir"); err != nil {
		ret.ServerDir = def.ServerDir
	}
	return ret
}

func dbSaveSiteConfig(dat *siteData) error {
	var err error
	if err = openDatabase(); err != nil {
		return err
	}
	defer closeDatabase()

	siteConf := []string{"site"}
	if err = db.SetValue(siteConf, "title", dat.Title); err != nil {
		return err
	}
	if err = db.SetInt(siteConf, "port", dat.Port); err != nil {
		return err
	}
	if err = db.SetValue(siteConf, "session-name", dat.SessionName); err != nil {
		return err
	}
	return db.SetValue(siteConf, "server-dir", dat.ServerDir)
}

func dbGetAuthMode() int {
	if ret, err := db.GetInt([]string{"site"}, "auth-mode"); err != nil {
		return AuthModeAuthentication
	} else {
		return ret
	}
}

func dbSetAuthMode(mode int) error {
	if mode < 0 || mode >= AuthModeError {
		return errors.New("Invalid site mode")
	}
	return db.SetInt([]string{"site"}, "auth-mode", mode)
}

func dbGetPublicSiteMode() int {
	if ret, err := db.GetInt([]string{"site"}, "public-mode"); err != nil {
		return SiteModeWaiting
	} else {
		return ret
	}
}

func dbSetPublicSiteMode(mode int) error {
	if mode < 0 || mode >= SiteModeError {
		return errors.New("Invalid site mode")
	}
	return db.SetInt([]string{"site"}, "public-mode", mode)
}
