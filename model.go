package main

import (
	"errors"
	"strings"

	"github.com/br0xen/boltease"
)

type gjDatabase struct {
	bolt     *boltease.DB
	dbOpened int
}

var db *gjDatabase

const (
	AuthModeAuthentication = iota
	AuthModeAll
	AuthModeError
)

func (db *gjDatabase) open() error {
	db.dbOpened += 1
	if db.dbOpened == 1 {
		var err error
		db.bolt, err = boltease.Create(DbName, 0600, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *gjDatabase) close() error {
	db.dbOpened -= 1
	if db.dbOpened == 0 {
		return db.bolt.CloseDB()
	}
	return nil
}

func (db *gjDatabase) initialize() error {
	var err error
	if err = db.open(); err != nil {
		return err
	}
	defer db.close()

	// Create the path to the bucket to store admin users
	if err := db.bolt.MkBucketPath([]string{"users"}); err != nil {
		return err
	}
	// Create the path to the bucket to store jam informations
	if err := db.bolt.MkBucketPath([]string{"jam"}); err != nil {
		return err
	}
	// Create the path to the bucket to store site config data
	return db.bolt.MkBucketPath([]string{"site"})
}

func (db *gjDatabase) setCurrentJam(name string) error {
	var err error
	if err = db.open(); err != nil {
		return err
	}
	defer db.close()

	return db.bolt.SetValue([]string{"site"}, "current-jam", name)
}

func (db *gjDatabase) hasCurrentJam() bool {
	var err error
	if _, err = db.getCurrentJam(); err != nil {
		return false
	}
	return true
}

func (db *gjDatabase) getCurrentJam() (string, error) {
	var ret string
	var err error
	if err = db.open(); err != nil {
		return "", err
	}
	defer db.close()

	ret, err = db.bolt.GetValue([]string{"site"}, "current-jam")

	if err == nil && strings.TrimSpace(ret) == "" {
		return ret, errors.New("No Jam Name Specified")
	}
	return ret, err
}

func (db *gjDatabase) getSiteConfig() *siteData {
	var ret *siteData
	def := NewSiteData()
	var err error
	if err = db.open(); err != nil {
		return def
	}
	defer db.close()

	ret = new(siteData)
	siteConf := []string{"site"}
	if ret.Title, err = db.bolt.GetValue(siteConf, "title"); err != nil {
		ret.Title = def.Title
	}
	if ret.Port, err = db.bolt.GetInt(siteConf, "port"); err != nil {
		ret.Port = def.Port
	}
	if ret.SessionName, err = db.bolt.GetValue(siteConf, "session-name"); err != nil {
		ret.SessionName = def.SessionName
	}
	if ret.ServerDir, err = db.bolt.GetValue(siteConf, "server-dir"); err != nil {
		ret.ServerDir = def.ServerDir
	}
	return ret
}

func (db *gjDatabase) getAuthMode() int {
	if ret, err := db.bolt.GetInt([]string{"site"}, "auth-mode"); err != nil {
		return AuthModeAuthentication
	} else {
		return ret
	}
}

func (db *gjDatabase) setAuthMode(mode int) error {
	if mode < 0 || mode >= AuthModeError {
		return errors.New("Invalid site mode")
	}
	return db.bolt.SetInt([]string{"site"}, "auth-mode", mode)
}

func (db *gjDatabase) getPublicSiteMode() int {
	if ret, err := db.bolt.GetInt([]string{"site"}, "public-mode"); err != nil {
		return SiteModeWaiting
	} else {
		return ret
	}
}

func dbSetPublicSiteMode(mode int) error {
	if mode < 0 || mode >= SiteModeError {
		return errors.New("Invalid site mode")
	}
	return db.bolt.SetInt([]string{"site"}, "public-mode", mode)
}
