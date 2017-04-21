package main

import "github.com/br0xen/boltease"

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
	if err := db.MkBucketPath([]string{"jams"}); err != nil {
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
	var nm string
	var err error
	if nm, err = dbGetCurrentJam(); err != nil {
		return false
	}
	ret, err := dbIsValidJam(nm)
	return ret && err != nil
}

func dbGetCurrentJam() (string, error) {
	if err := db.OpenDB(); err != nil {
		return "", err
	}
	defer db.CloseDB()

	return db.GetValue([]string{"site"}, "current-jam")
}

func dbIsValidJam(name string) (bool, error) {
	var err error
	if err = db.OpenDB(); err != nil {
		return false, err
	}
	defer db.CloseDB()

	// Get all keys in the jams bucket
	var keys []string
	if keys, err = db.GetKeyList([]string{"jams", name}); err != nil {
		return false, err
	}
	// All valid gamejams will have:
	//	"name"
	//	"teams"
	for _, v := range []string{"name", "teams"} {
		found := false
		for j := range keys {
			if keys[j] == v {
				found = true
				break
			}
		}
		if !found {
			// If we make it here, we didn't find a key we need
			return false, nil
		}
	}
	return true, nil
}
