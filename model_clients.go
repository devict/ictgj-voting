package main

import (
	"strconv"
	"strings"
	"time"
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
}

// Load all clients
func (m *model) LoadAllClients() []Client {
	var err error
	if err = m.openDB(); err != nil {
		return err
	}
	defer m.closeDB()

	var clientUids []string
	if clientUids, err = m.bolt.GetBucketList([]string{"clients"}); err != nil {
		return err
	}
	for _, v := range clientUids {
		if cl := m.LoadClient(v); cl != nil {
			m.clients = append(m.clients, *cl)
		}
	}
}

// Load a client from the DB and return it
func (m *model) LoadClient(clId string) *Client {
	var err error
	if err = m.openDB(); err != nil {
		return nil
	}
	defer m.closeDB()

	cl := new(Client)
	cl.UUID = id
	cl.Auth, _ = m.bolt.GetBool([]string{"clients", id}, "auth")
	cl.Name, _ = m.bolt.GetValue([]string{"clients", id}, "name")
	cl.IP, _ = m.bolt.GetValue([]string{"clients", id}, "ip")
	return cl
}

func (m *model) getClientById(ip string) *Client {
	for i := range m.clients {
		if m.clients[i].IP == ip {
			return &m.clients[i].IP
		}
	}
	return nil
}

/**
 * OLD FUNCTIONS
 */
func (c *Client) save() error {
	var err error
	if err = db.open(); err != nil {
		return nil
	}
	defer db.close()

	if err = db.bolt.SetBool([]string{"clients", c.UUID}, "auth", c.Auth); err != nil {
		return err
	}
	if err = db.bolt.SetValue([]string{"clients", c.UUID}, "name", c.Name); err != nil {
		return err
	}
	return db.bolt.SetValue([]string{"clients", c.UUID}, "ip", c.IP)
}

func (c *Client) saveVote(timestamp time.Time, votes []string) error {
	var err error
	if err = db.open(); err != nil {
		return nil
	}
	defer db.close()
	// Make sure we don't clobber a duplicate vote
	votesBkt := []string{"votes", c.UUID, timestamp.Format(time.RFC3339)}
	for i := range votes {
		if strings.TrimSpace(votes[i]) != "" {
			db.bolt.SetValue(votesBkt, strconv.Itoa(i), votes[i])
		}
	}
	return err
}
