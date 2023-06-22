/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package provisioner

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Session stores a TPM Provisioner session data.
type Session struct {
	xname    string
	nodeType string
	expiry   time.Time
	step     int
	nonce    string
	reqData  string
}

var sessions = map[string]Session{}

// createSessionCookie returns a basic session cookie.
func createSessionCookie(xname string, nodeType string) (string, time.Time) {
	token := uuid.NewString()
	expiresAt := time.Now().Add(2 * time.Minute)

	sessions[token] = Session{
		xname:    xname,
		nodeType: nodeType,
		expiry:   expiresAt,
		step:     0,
	}

	return token, expiresAt
}

// CleanSessions looks for expired sessions and removes them.
func CleanSessions() {
	for k, v := range sessions {
		if v.expiry.Before(time.Now()) {
			delete(sessions, k)
		}
	}

	time.Sleep(2 * time.Minute)
}

// validateCookie validates that a session cookie exists and that it's being
// used in the correct step.
func validateCookie(c []*http.Cookie, step int) error {
	sessionCookie, err := getSession(c)
	if err != nil {
		return err
	}

	session := sessions[sessionCookie]

	if (session == Session{}) {
		return errors.New("invalid session cookie")
	}

	if session.step != step {
		return errors.New("request out of order")
	}

	if session.expiry.Before(time.Now()) {
		return errors.New("session expired")
	}

	// Increment the session step so that a step can not be run twice or run
	// out of order
	session.step += 1
	sessions[sessionCookie] = session

	return nil
}

// getSession returns a session associated with the session cookie.
func getSession(c []*http.Cookie) (string, error) {
	var sessionCookie string

	if len(c) == 0 {
		return "", errors.New("missing session cookie")
	}

	for _, v := range c {
		b := false

		if v.Name == "session" {
			sessionCookie = v.Value
			b = true
		}

		if !b {
			return "", errors.New("missing session cookie")
		}
	}
	return sessionCookie, nil
}

// setNonce adds the validation nonce to the session.
func setNonce(c []*http.Cookie, nonce string) error {
	sessionCookie, err := getSession(c)
	if err != nil {
		return err
	}

	session := sessions[sessionCookie]
	session.nonce = nonce
	sessions[sessionCookie] = session

	return nil
}

// getReqData returns the reqData from the session associated with the session
// cookie.
func getReqData(c []*http.Cookie) (string, error) {
	sessionCookie, err := getSession(c)
	if err != nil {
		return "", err
	}

	session := sessions[sessionCookie]

	return session.reqData, nil
}

// setReqData sets the reqData from the session associated with the session
// cookie.
func setReqData(c []*http.Cookie, reqData string) error {
	sessionCookie, err := getSession(c)
	if err != nil {
		return err
	}

	session := sessions[sessionCookie]
	session.reqData = reqData
	sessions[sessionCookie] = session

	return nil
}

// getNonce gets the nonce from the session associated with the session cookie.
func getNonce(c []*http.Cookie) (string, error) {
	sessionCookie, err := getSession(c)
	if err != nil {
		return "", err
	}

	session := sessions[sessionCookie]

	return session.nonce, nil
}

// getXname gets the xname from the session associated with the session cookie.
func getXname(c []*http.Cookie) (string, error) {
	sessionCookie, err := getSession(c)
	if err != nil {
		return "", err
	}

	session := sessions[sessionCookie]

	return session.xname, nil
}

// getType gets the node type from the session associated with the session cookie.
func getType(c []*http.Cookie) (string, error) {
	sessionCookie, err := getSession(c)
	if err != nil {
		return "", err
	}

	session := sessions[sessionCookie]

	return session.nodeType, nil
}
