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
// tpm-provisioner-client requests a signed devid from the tpm-provisioner
// server.
// by default the config file is /etc/tpm-provisioner/client.conf
package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/cray-hpe/tpm-provisioner/pkg/client"
	"github.com/google/go-tpm/legacy/tpm2"
)

func main() {
	ctx := context.Background()

	if len(os.Args) > 2 {
		log.Fatalf("%s [CONFIG FILE]", os.Args[0])
	}

	var f string

	var p string

	if len(os.Args) == 1 {
		p = "/etc/tpm-provisioner"
		f = "client.conf"
	} else {
		s := strings.Split(os.Args[1], "/")
		p = strings.Join(s[:len(s)-1], "/")
		f = s[len(s)-1]
	}

	cfg, err := client.ParseConfig(p, f)
	if err != nil {
		log.Fatalf("Unable to parse config: %v", err)
	}

	rwc, err := tpm2.OpenTPM("/dev/tpmrm0")
	if err != nil {
		log.Fatalf("Error opening TPM: %v", err)
	}

	defer func() {
		if err = rwc.Close(); err != nil {
			log.Fatalf("Error closing TPM: %v", err)
		}
	}()

	id, err := getIdentity()
	if err != nil {
		log.Printf("Failed to get identity: %v", err)
		return
	}

	var jwt string

	if cfg.SocketPath != "" {
		jwt, err = client.FetchJWT(cfg.SocketPath)
		if err != nil {
			log.Printf("Unable to get JWT from Spire: %v", err)
			return
		}
	}

	sessionCookie, err := authorize(id, cfg.URL, jwt)
	if err != nil {
		log.Printf("authorization failed: %v", err)
		return
	}

	requestData, requestSig, resources, err := client.CreateRawRequest(ctx, rwc, id)
	if err != nil {
		log.Printf("creating raw request failed: %v", err)
		return
	}

	cResp, err := challengeRequest(requestData, requestSig, sessionCookie, cfg.URL, jwt)
	if err != nil {
		log.Printf("challenge request failed: %v", err)
		return
	}

	cSubmit, err := client.GenerateChallengeResponse(rwc, cResp.Blob, cResp.Secret, resources)
	if err != nil {
		log.Printf("generate challenge response failed: %v", err)
		return
	}

	devID, err := challengeSubmit(cSubmit, sessionCookie, cfg.URL, jwt)
	if err != nil {
		log.Printf("challenge submission failed: %v", err)
		return
	}

	err = client.WriteDevID(cfg.OutputDir, resources, devID)
	if err != nil {
		log.Printf("failed to write blobs to %s: %v", cfg.OutputDir, err)
		return
	}
}
