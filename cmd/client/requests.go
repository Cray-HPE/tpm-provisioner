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
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/cray-hpe/tpm-provisioner/pkg/provisioner"
)

// authorize requests a session cookie from the tpm-provisioning server.
func authorize(id pkix.Name, url string, jwt string) (string, error) {
	nodeType := strings.Split(id.CommonName, "/")[0]
	xname := strings.Split(id.CommonName, "/")[1]

	httpClient := http.Client{}

	log.Printf("url: %s", url)
	log.Printf("xname: %s", xname)
	log.Printf("type: %s", nodeType)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/authorize?xname=%s&type=%s", url, xname, nodeType), nil)
	log.Printf("req: %+v", req)

	if err != nil {
		return "", err
	}

	if jwt != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("error authorizing to tpm-provisioner: %+v: %v", resp.StatusCode, string(data))
	}

	var j provisioner.AuthorizeResponse

	err = json.NewDecoder(resp.Body).Decode(&j)
	if err != nil {
		return "", err
	}

	err = resp.Body.Close()
	if err != nil {
		return "", err
	}

	if !j.Success {
		log.Fatalf("Received Failure Reason: %v", j.Reason)
	}

	if len(resp.Cookies()) == 0 {
		log.Fatalf("No Cookies set")
	}

	var sessionCookie string

	for _, v := range resp.Cookies() {
		b := false
		if v.Name == "session" {
			b = true
			sessionCookie = v.Value
		}

		if !b {
			log.Fatalf("Session Cookie Not Found.")
		}
	}

	return sessionCookie, nil
}

// challengeRequest sends a challenge request to the tpm-provisioner server.
func challengeRequest(data []byte, sig []byte, sessionCookie string, url string, jwt string) (provisioner.CertificateResponse, error) {
	reqData := provisioner.CertificateRequest{
		Data: base64.StdEncoding.EncodeToString(data),
		Sig:  base64.StdEncoding.EncodeToString(sig),
	}

	body, err := json.Marshal(reqData)
	if err != nil {
		log.Fatalf("Failed to Marshall json: %v", err)
	}

	httpClient := http.Client{}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/challenge/request", url), bytes.NewBuffer(body))
	if err != nil {
		return provisioner.CertificateResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	if jwt != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
	}

	var certResp provisioner.CertificateResponse

	req.AddCookie(&http.Cookie{Name: "session", Value: sessionCookie})

	resp, err := httpClient.Do(req)
	if err != nil {
		return provisioner.CertificateResponse{}, err
	}

	err = json.NewDecoder(resp.Body).Decode(&certResp)
	if err != nil {
		return provisioner.CertificateResponse{}, err
	}

	err = resp.Body.Close()
	if err != nil {
		return provisioner.CertificateResponse{}, err
	}

	if !certResp.Success {
		log.Fatalf("Failed to request challenge: %v", certResp.Reason)
	}

	return certResp, nil
}

// challengeSubmit submits the challenge response to the tpm-provisioner server.
func challengeSubmit(data []byte, sessionCookie string, url string, jwt string) ([]byte, error) {
	submission := provisioner.SubmitRequest{
		Data: base64.StdEncoding.EncodeToString(data),
	}

	httpClient := http.Client{}

	body, err := json.Marshal(submission)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/challenge/submit", url), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	req.AddCookie(&http.Cookie{Name: "session", Value: sessionCookie})

	if jwt != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	var submitResp provisioner.SubmitResponse

	err = json.NewDecoder(resp.Body).Decode(&submitResp)
	if err != nil {
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	if !submitResp.Success {
		log.Fatalf("Failed to request challenge: %v", submitResp.Reason)
	}

	devID, err := base64.RawStdEncoding.DecodeString(submitResp.DevIDCertificate)
	if err != nil {
		return nil, err
	}

	return devID, nil
}
