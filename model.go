package main

import (
	"errors"
	"strings"

	"github.com/br0xen/boltease"
)

var db *boltease.DB
var dbOpened bool

func openDatabase() error {
	if !dbOpened {
		var err error
		db, err = boltease.Create(site.DB, 0600, nil)
		if err != nil {
			return err
		}
		dbOpened = true
	}
	return nil
}

func initDatabase() error {
	openDatabase()
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
	if err := db.OpenDB(); err != nil {
		return err
	}
	defer db.CloseDB()

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
	if err = db.OpenDB(); err != nil {
		return "", err
	}
	defer db.CloseDB()

	ret, err = db.GetValue([]string{"site"}, "current-jam")

	if err == nil && strings.TrimSpace(ret) == "" {
		return ret, errors.New("No Jam Name Specified")
	}
	return ret, err
}
