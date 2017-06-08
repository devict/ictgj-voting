package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/pborman/uuid"
)

// This is basically a convenience struct for
// easier session management (hopefully ;)
type pageSession struct {
	session *sessions.Session
	req     *http.Request
	w       http.ResponseWriter
}

func (p *pageSession) getStringValue(key string) (string, error) {
	val := p.session.Values[key]
	var retVal string
	var ok bool
	if retVal, ok = val.(string); !ok {
		return "", fmt.Errorf("Unable to create string from %s", key)
	}
	return retVal, nil
}

func (p *pageSession) setStringValue(key, val string) {
	p.session.Values[key] = val
	p.session.Save(p.req, p.w)
}

func (p *pageSession) getClientID() string {
	var clientId string
	var err error
	if clientId, err = p.getStringValue("client_id"); err != nil {
		// No client id, generate and save one
		clientId := uuid.New()
		p.setStringValue("client_id", clientId)
	}
	return clientId
}

func (p *pageSession) setFlashMessage(msg, status string) {
	p.setStringValue("flash_message", msg)
	p.setStringValue("flash_status", status)
}

func (p *pageSession) getFlashMessage() (string, string) {
	var err error
	var msg, status string
	if msg, err = p.getStringValue("flash_message"); err != nil {
		return "", "hidden"
	}
	if status, err = p.getStringValue("flash_status"); err != nil {
		return "", "hidden"
	}
	p.setFlashMessage("", "hidden")
	return msg, status
}

func (p *pageSession) expireSession() {
	p.session.Options.MaxAge = -1
	p.session.Save(p.req, p.w)
}
