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
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cray-hpe/tpm-provisioner/pkg/client"
	"github.com/cray-hpe/tpm-provisioner/pkg/provisioner"
	"github.com/cray-hpe/tpm-provisioner/tests/simulateTPM"
	"github.com/google/go-tpm-tools/simulator"
)

// openTestTPM opens the tpm simulator.
func openTestTPM(tb testing.TB) io.ReadWriteCloser {
	tb.Helper()

	simulator, err := simulator.Get()
	if err != nil {
		tb.Fatalf("Simulator initialization failed: %v", err)
	}

	return simulator
}

// TestHappyPath validates that the entire DevID sign request process works
// properly.
func TestHappyPath(t *testing.T) {
	ctx := context.Background()

	rwc := openTestTPM(t)

	caCRT, err := simulateTPM.CreateEK(rwc)
	if err != nil {
		t.Fatalf("Unable to create EK: %v", err)
	}

	certPool := x509.NewCertPool()

	if !certPool.AppendCertsFromPEM(caCRT) {
		t.Fatalf("Unable to Add CA to cert pool")
	}

	pCA, pPrivKey, _, err := simulateTPM.GenerateCA("Provisioner CA")
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

	provisioner.WhiteList = append(provisioner.WhiteList, "x1000c0s0b0n0")

	id := pkix.Name{
		CommonName: "compute/x1000c0s0b0n0",
	}

	router := provisioner.NewRouter()

	ts := httptest.NewServer(router)

	tsURL := ts.URL + "/apis/tpm-provisioner"

	sessionCookie, err := authorize(id, tsURL, "")
	if err != nil {
		t.Fatalf("authorization failed: %v", err)
	}

	requestData, requestSig, resources, err := client.CreateRawRequest(ctx, rwc, id)
	if err != nil {
		t.Fatalf("creating raw request failed: %v", err)
	}

	cResp, err := challengeRequest(requestData, requestSig, sessionCookie, tsURL, "")
	if err != nil {
		t.Fatalf("challenge request failed: %v", err)
	}

	cSubmit, err := client.GenerateChallengeResponse(rwc, cResp.Blob, cResp.Secret, resources)
	if err != nil {
		t.Fatalf("generate challenge response failed: %v", err)
	}

	_, err = challengeSubmit(cSubmit, sessionCookie, tsURL, "")
	if err != nil {
		t.Fatalf("challenge submission failed: %v", err)
	}

	ts.Close()
}
