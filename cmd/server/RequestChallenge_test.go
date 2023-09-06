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
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cray-hpe/tpm-provisioner/pkg/provisioner"
)

// TestRequestChallengeWithInvalidSessionCookie validates that using an invalid
// session cookie fails.
func TestRequestChallengeWithInvalidSessionCookie(t *testing.T) {
	rr := httptest.NewRecorder()

	r := strings.NewReader("certrequest")

	req, err := http.NewRequest("POST", "/api/request", r)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	req.AddCookie(&http.Cookie{
		Name:    "session",
		Value:   "0abc91c1-e36d-42b8-a71b-3e64263f5f92",
		Expires: time.Now().Add(2 * time.Minute),
	})

	handler := http.HandlerFunc(provisioner.RequestChallenge)
	handler.ServeHTTP(rr, req)

	var resp provisioner.CertificateResponse

	err = json.NewDecoder(rr.Body).Decode(&resp)
	if err != nil {
		t.Fatal(err)
	}

	// Expect Success to be true
	if resp.Success != false {
		t.Fail()
	}

	expected := "invalid session cookie"
	if resp.Reason != expected {
		t.Fatalf("Expected: %s\nReceived: %s", expected, resp.Reason)
	}
}

// TestRequestChallengeWithoutSessionCookie validates that it's not possible to
//
//	request a challenge without a session cookie
func TestRequestChallengeWithoutSessionCookie(t *testing.T) {
	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()

	// Send Authorization Request
	r := strings.NewReader("certrequest")

	req, err := http.NewRequest("POST", "/api/request", r)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler := http.HandlerFunc(provisioner.RequestChallenge)
	handler.ServeHTTP(rr, req)

	var resp provisioner.AuthorizeResponse

	err = json.NewDecoder(rr.Body).Decode(&resp)
	if err != nil {
		t.Fatal(err)
	}

	// Expect Success to be true
	if resp.Success != false {
		t.Fail()
	}

	expected := "missing session cookie"
	if resp.Reason != expected {
		t.Fatalf("Expected: %s\nReceived: %s", expected, resp.Reason)
	}
}

// TestRequestChallengeWithoutNonSessionCookie validates that it's not possible
// to request a challenge with an invalid session cookie.
func TestRequestChallengeWithoutNonSessionCookie(t *testing.T) {
	rr := httptest.NewRecorder()

	r := strings.NewReader("certrequest")

	req, err := http.NewRequest("POST", "/api/request", r)
	if err != nil {
		t.Fatal(err)
	}

	req.AddCookie(&http.Cookie{
		Name:    "junk",
		Value:   "junk",
		Expires: time.Now().Add(2 * time.Hour),
	})

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler := http.HandlerFunc(provisioner.RequestChallenge)
	handler.ServeHTTP(rr, req)

	var resp provisioner.AuthorizeResponse

	err = json.NewDecoder(rr.Body).Decode(&resp)
	if err != nil {
		t.Fatal(err)
	}

	// Expect Success to be true
	if resp.Success != false {
		t.Fail()
	}

	expected := "missing session cookie"
	if resp.Reason != expected {
		t.Fatalf("Expected: %s\nReceived: %s", expected, resp.Reason)
	}
}
