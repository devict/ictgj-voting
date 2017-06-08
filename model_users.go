package main

import "golang.org/x/crypto/bcrypt"

// dbHasUser
// Returns true if there are any users in the database
func dbHasUser() bool {
	return len(dbGetAllUsers()) > 0
}

func dbGetAllUsers() []string {
	if err := db.OpenDB(); err != nil {
		return []string{}
	}
	defer db.CloseDB()

	usrs, err := db.GetBucketList([]string{"users"})
	if err != nil {
		return []string{}
	}
	return usrs
}

func dbIsValidUserEmail(email string) bool {
	if err := db.OpenDB(); err != nil {
		return false
	}
	defer db.CloseDB()

	usrPath := []string{"users", email}
	_, err := db.GetValue(usrPath, "password")
	return err == nil
}

func dbCheckCredentials(email, pw string) error {
	var err error
	if err = db.OpenDB(); err != nil {
		return err
	}
	defer db.CloseDB()

	var uPw string
	usrPath := []string{"users", email}
	if uPw, err = db.GetValue(usrPath, "password"); err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword([]byte(uPw), []byte(pw))
}

// dbUpdateUserPassword
// Takes an email address and a password
// Creates the user if it doesn't exist, encrypts the password
// and updates it in the db
func dbUpdateUserPassword(email, password string) error {
	cryptPw, cryptError := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if cryptError != nil {
		return cryptError
	}
	if err := db.OpenDB(); err != nil {
		return err
	}
	defer db.CloseDB()

	usrPath := []string{"users", email}
	return db.SetValue(usrPath, "password", string(cryptPw))
}

func dbDeleteUser(email string) error {
	var err error
	if err = db.OpenDB(); err != nil {
		return err
	}
	defer db.CloseDB()

	return db.DeleteBucket([]string{"users"}, email)
}
