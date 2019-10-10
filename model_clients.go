package main

import (
	"errors"

	"github.com/pborman/uuid"
)

/**
 * Client
 * A client is a system that is connecting to the web server
 */
type Client struct {
	UUID string
	Auth bool
	Name string
	IP   string

	mPath []string // The path in the DB to this client
}

func NewClient(id string) *Client {
	if id == "" {
		id = uuid.New()
	}
	return &Client{
		UUID:  id,
		mPath: []string{"clients", id},
	}
}

func (m *model) AddClient(cl *Client) error {
	for i := range m.clients {
		if m.clients[i].UUID == cl.UUID {
			return errors.New("A client with that ID already exists")
		}
		if m.clients[i].IP == cl.IP {
			return errors.New("A client with that IP already exists")
		}
		if m.clients[i].Name == cl.Name {
			return errors.New("A client with that Name already exists")
		}
	}
	m.clients = append(m.clients, *cl)
	m.clientsUpdated = true
	return nil
}

/**
 * DB Functions
 * These are generally just called when the app starts up, or when the periodic 'save' runs
 */

// Load all clients from the DB
func (m *model) LoadAllClients() []Client {
	var err error
	var ret []Client
	if err = m.openDB(); err != nil {
		return ret
	}
	defer m.closeDB()

	var clientUids []string
	cliPath := []string{"clients"}
	if clientUids, err = m.bolt.GetBucketList(cliPath); err != nil {
		return ret
	}
	for _, v := range clientUids {
		if cl := m.LoadClient(v); cl != nil {
			ret = append(ret, *cl)
		}
	}
	return ret
}

// Load a client from the DB and return it
func (m *model) LoadClient(clId string) *Client {
	var err error
	if err = m.openDB(); err != nil {
		return nil
	}
	defer m.closeDB()

	cl := NewClient(clId)
	cl.Auth, _ = m.bolt.GetBool(cl.mPath, "auth")
	cl.Name, _ = m.bolt.GetValue(cl.mPath, "name")
	cl.IP, _ = m.bolt.GetValue(cl.mPath, "ip")
	return cl
}

// SaveAllClients saves all clients to the DB
func (m *model) SaveAllClients() error {
	var err error
	if err = m.openDB(); err != nil {
		return nil
	}
	defer m.closeDB()

	for _, v := range m.clients {
		if err = m.SaveClient(&v); err != nil {
			return err
		}
	}
	return nil
}

// SaveClient saves a client to the DB
func (m *model) SaveClient(cl *Client) error {
	var err error
	if err = m.openDB(); err != nil {
		return nil
	}
	defer m.closeDB()

	if err = m.bolt.SetBool(cl.mPath, "auth", cl.Auth); err != nil {
		return err
	}
	if err = m.bolt.SetValue(cl.mPath, "name", cl.Name); err != nil {
		return err
	}
	return m.bolt.SetValue(cl.mPath, "ip", cl.IP)
}

/**
 * In Memory functions
 * This is generally how the app accesses client data
 */

// Return a client by it's UUID
func (m *model) GetClient(id string) (*Client, error) {
	for i := range m.clients {
		if m.clients[i].UUID == id {
			return &m.clients[i], nil
		}
	}
	return nil, errors.New("Invalid Id")
}

// Return a client by it's IP address
func (m *model) GetClientByIp(ip string) (*Client, error) {
	for i := range m.clients {
		if m.clients[i].IP == ip {
			return &m.clients[i], nil
		}
	}
	return nil, errors.New("Invalid Ip")
}

// Add/Update a client in the data model
func (m *model) UpdateClient(cl *Client) {
	var found bool
	for i := range m.clients {
		if m.clients[i].UUID == cl.UUID {
			found = true
			m.clients[i].Auth = cl.Auth
			m.clients[i].Name = cl.Name
			m.clients[i].IP = cl.IP
		}
	}
	if !found {
		m.clients = append(m.clients, *cl)
	}
	m.clientsUpdated = true
}

func (m *model) DeleteClient(id string) error {
	idx := -1
	for i := range m.clients {
		if m.clients[i].UUID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return errors.New("Invalid Client ID given")
	}
	m.clients = append(m.clients[:idx], m.clients[idx+1:]...)
	// Now delete it from the DB
	if err := m.openDB(); err != nil {
		return nil
	}
	defer m.closeDB()
	return m.bolt.DeleteBucket([]string{"clients"}, id)
}
