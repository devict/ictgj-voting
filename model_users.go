package main

import "golang.org/x/crypto/bcrypt"

// dbHasUser
// Returns true if there are any users in the database
func (db *gjDatabase) hasUser() bool {
	return len(db.getAllUsers()) > 0
}

func (db *gjDatabase) getAllUsers() []string {
	if err := db.open(); err != nil {
		return []string{}
	}
	defer db.close()

	usrs, err := db.bolt.GetBucketList([]string{"users"})
	if err != nil {
		return []string{}
	}
	return usrs
}

func (db *gjDatabase) isValidUserEmail(email string) bool {
	if err := db.open(); err != nil {
		return false
	}
	defer db.close()

	usrPath := []string{"users", email}
	_, err := db.bolt.GetValue(usrPath, "password")
	return err == nil
}

func (db *gjDatabase) checkCredentials(email, pw string) error {
	var err error
	if err = db.open(); err != nil {
		return err
	}
	defer db.close()

	var uPw string
	usrPath := []string{"users", email}
	if uPw, err = db.bolt.GetValue(usrPath, "password"); err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword([]byte(uPw), []byte(pw))
}

// updateUserPassword
// Takes an email address and a password
// Creates the user if it doesn't exist, encrypts the password
// and updates it in the db
func (db *gjDatabase) updateUserPassword(email, password string) error {
	cryptPw, cryptError := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if cryptError != nil {
		return cryptError
	}
	if err := db.open(); err != nil {
		return err
	}
	defer db.close()

	usrPath := []string{"users", email}
	return db.bolt.SetValue(usrPath, "password", string(cryptPw))
}

func (db *gjDatabase) deleteUser(email string) error {
	var err error
	if err = db.open(); err != nil {
		return err
	}
	defer db.close()

	return db.bolt.DeleteBucket([]string{"users"}, email)
}
