package main

import "golang.org/x/crypto/bcrypt"

// These are all model functions that have to do with users

// Returns true if there are any users in the database
func (m *model) hasUser() bool {
	return len(m.getAllUsers()) > 0
}

func (m *model) getAllUsers() []string {
	if err := m.openDB(); err != nil {
		return []string{}
	}
	defer m.closeDB()

	usrs, err := m.bolt.GetBucketList([]string{"users"})
	if err != nil {
		return []string{}
	}
	return usrs
}

// Is the given email one that is in our DB?
func (m *model) isValidUserEmail(email string) bool {
	if err := m.openDB(); err != nil {
		return false
	}
	defer m.closeDB()

	usrPath := []string{"users", email}
	_, err := m.bolt.GetValue(usrPath, "password")
	return err == nil
}

// Is the email and pw given valid?
func (m *model) checkCredentials(email, pw string) error {
	var err error
	if err = m.openDB(); err != nil {
		return err
	}
	defer m.closeDB()

	var uPw string
	usrPath := []string{"users", email}
	if uPw, err = m.bolt.GetValue(usrPath, "password"); err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword([]byte(uPw), []byte(pw))
}

// updateUserPassword
// Takes an email address and a password
// Creates the user if it doesn't exist, encrypts the password
// and updates it in the db
func (m *model) updateUserPassword(email, password string) error {
	cryptPw, cryptError := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if cryptError != nil {
		return cryptError
	}
	if err := m.openDB(); err != nil {
		return err
	}
	defer m.closeDB()

	usrPath := []string{"users", email}
	return m.bolt.SetValue(usrPath, "password", string(cryptPw))
}

func (m *model) deleteUser(email string) error {
	var err error
	if err = m.openDB(); err != nil {
		return err
	}
	defer m.closeDB()

	return m.bolt.DeleteBucket([]string{"users"}, email)
}
