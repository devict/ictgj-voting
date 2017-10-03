package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	UUID string
	Auth bool
	Name string
	IP   string
}

func (db *gjDatabase) getAllClients() []Client {
	var ret []Client
	var err error
	if err = db.open(); err != nil {
		return ret
	}
	defer db.close()

	var clientUids []string
	if clientUids, err = db.bolt.GetBucketList([]string{"clients"}); err != nil {
		return ret
	}
	for _, v := range clientUids {
		if cl := db.getClient(v); cl != nil {
			ret = append(ret, *cl)
		}
	}
	return ret
}

func (db *gjDatabase) getClient(id string) *Client {
	var err error
	if err = db.open(); err != nil {
		return nil
	}
	defer db.close()

	cl := new(Client)
	cl.UUID = id
	cl.Auth, _ = db.bolt.GetBool([]string{"clients", id}, "auth")
	cl.Name, _ = db.bolt.GetValue([]string{"clients", id}, "name")
	cl.IP, _ = db.bolt.GetValue([]string{"clients", id}, "ip")
	return cl
}

func (db *gjDatabase) getClientByIp(ip string) *Client {
	var err error
	if err = db.open(); err != nil {
		return nil
	}
	defer db.close()

	allClients := db.getAllClients()
	for i := range allClients {
		if allClients[i].IP == ip {
			return &allClients[i]
		}
	}
	return nil
}

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

func (c *Client) getVotes() []Vote {
	var ret []Vote
	var err error
	if err = db.open(); err != nil {
		return ret
	}
	defer db.close()

	var times []string
	votesBkt := []string{"votes", c.UUID}
	if times, err = db.bolt.GetBucketList(votesBkt); err != nil {
		return ret
	}
	for _, t := range times {
		var tm time.Time
		if tm, err = time.Parse(time.RFC3339, t); err == nil {
			var vt *Vote
			if vt, err = c.getVote(tm); err == nil {
				ret = append(ret, *vt)
			} else {
				fmt.Println(err)
			}
		}
	}
	return ret
}

func (c *Client) getVote(timestamp time.Time) (*Vote, error) {
	var err error
	if err = db.open(); err != nil {
		return nil, err
	}
	defer db.close()

	vt := new(Vote)
	vt.Timestamp = timestamp
	vt.ClientId = c.UUID
	votesBkt := []string{"votes", c.UUID, timestamp.Format(time.RFC3339)}
	var choices []string
	if choices, err = db.bolt.GetKeyList(votesBkt); err != nil {
		// Couldn't find the vote...
		return nil, err
	}
	for _, v := range choices {
		ch := new(GameChoice)
		var rank int

		if rank, err = strconv.Atoi(v); err == nil {
			ch.Rank = rank
			ch.Team, err = db.bolt.GetValue(votesBkt, v)
			vt.Choices = append(vt.Choices, *ch)
		}
	}
	return vt, nil
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
