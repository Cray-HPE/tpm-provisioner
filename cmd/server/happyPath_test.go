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
	"bytes"
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cray-hpe/tpm-provisioner/pkg/client"
	"github.com/cray-hpe/tpm-provisioner/pkg/provisioner"
	"github.com/cray-hpe/tpm-provisioner/pkg/test"
	"github.com/google/go-tpm-tools/simulator"
)

func openTPM(tb testing.TB) io.ReadWriteCloser {
	tb.Helper()

	simulator, err := simulator.Get()
	if err != nil {
		tb.Fatalf("Simulator initialization failed: %v", err)
	}

	return simulator
}

// TestHappyPath validates that the whole server TPM provisioner process works.
func TestHappyPath(t *testing.T) {
	rw := openTPM(t)

	caCRT, err := test.CreateEK(rw)

	certPool := x509.NewCertPool()

	if !certPool.AppendCertsFromPEM(caCRT) {
		t.Fatalf("Unable to Add CA to cert pool")
	}

	if err != nil {
		t.Fatalf("Unable to provision EK: %v", err)
	}

	pCA, pPrivKey, _, err := test.GenerateCA("Provisioner CA")
	if err != nil {
		t.Fatalf("Unable to provision CA: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(`{"success":"true"}`))
		if err != nil {
			t.Fatalf("Unable to write response: %v", err)
		}
	}))

	defer server.Close()

	provisioner.CFG = provisioner.Config{
		ManufactuerCAs: certPool,
		ProviderCA:     pCA,
		ProviderKey:    pPrivKey,
		Port:           8080,
		WhiteList:      "/tmp/whitelist.tpm",
		SpireTokensURL: server.URL,
	}

	rr := httptest.NewRecorder()

	provisioner.WhiteList = append(provisioner.WhiteList, "x1000c0s0b0n0")

	r := strings.NewReader("")

	req, err := http.NewRequest("GET", "/api/authorize?xname=x1000c0s0b0n0&type=compute", r)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler := http.HandlerFunc(provisioner.Authorize)
	handler.ServeHTTP(rr, req)

	var resp provisioner.AuthorizeResponse

	err = json.NewDecoder(rr.Body).Decode(&resp)
	if err != nil {
		t.Fatal(err)
	}

	// Expect Success to be true
	if resp.Success != true {
		t.Fatalf("Received Failure Reason: %v", resp.Reason)
	}

	// Expect a Session cookie to be set
	if len(rr.Result().Cookies()) == 0 {
		t.Fatalf("No Cookies set")
	}

	var sessionCookie string

	for _, v := range rr.Result().Cookies() {
		b := false
		if v.Name == "session" {
			b = true
			sessionCookie = v.Value
		}

		if !b {
			t.Fatalf("Session Cookie Not Found.")
		}
	}

	ctx := context.Background()

	PlatformIdentity := pkix.Name{
		CommonName: "compute/x1000c0s0b0n0",
	}

	requestData, requestSig, resources, err := client.CreateRawRequest(ctx, rw, PlatformIdentity)
	if err != nil {
		t.Fatalf("Failed to Create Request: %v", err)
	}

	data := provisioner.CertificateRequest{
		Data: base64.StdEncoding.EncodeToString(requestData),
		Sig:  base64.StdEncoding.EncodeToString(requestSig),
	}

	body, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to Marshall json: %v", err)
	}

	if err != nil {
		t.Fatalf("Failed to Create Raw Request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	var certResp provisioner.CertificateResponse

	req, err = http.NewRequest("POST", "/api/challenge/request", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	req.AddCookie(&http.Cookie{Name: "session", Value: sessionCookie})

	certHandler := http.HandlerFunc(provisioner.RequestChallenge)
	certHandler.ServeHTTP(rr, req)

	err = json.NewDecoder(rr.Body).Decode(&certResp)
	if err != nil {
		t.Fatal(err)
	}

	if certResp.Success == false {
		t.Fatalf("Failed to request challenge: %v", certResp.Reason)
	}

	challengeResponse, err := client.GenerateChallengeResponse(rw, certResp.Blob, certResp.Secret, resources)
	if err != nil {
		t.Fatalf("Failed to Generate a challenge response: %v", err)
	}

	submission := provisioner.SubmitRequest{
		Data: base64.StdEncoding.EncodeToString(challengeResponse),
	}

	body, err = json.Marshal(submission)
	if err != nil {
		t.Fatalf("Failed to Marshall json: %v", err)
	}

	if err != nil {
		t.Fatalf("Failed to Create Raw Request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	var submitResponse provisioner.SubmitResponse

	req, err = http.NewRequest("POST", "/api/challenge/submit", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	req.AddCookie(&http.Cookie{Name: "session", Value: sessionCookie})

	certHandler = http.HandlerFunc(provisioner.SubmitChallenge)
	certHandler.ServeHTTP(rr, req)

	err = json.NewDecoder(rr.Body).Decode(&submitResponse)
	if err != nil {
		t.Fatal(err)
	}

	if submitResponse.Success == false {
		t.Fatalf("failed submitResponse: %#+v\n", submitResponse)
	}
}
