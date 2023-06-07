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
// Package main implements a simple way to retrieve the TPM's EK certificate.
package main

import (
	"bytes"
	"encoding/pem"
	"fmt"
	"io"
	"log"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// getEKPublicCertificate retrieves the EK Certificate from TPM.
func getEKPublicCertificate(rwc io.ReadWriter) (string, error) {
	const EKRSACertificateHandle = tpmutil.Handle(0x01c00002)

	ekCertData, err := tpm2.NVRead(rwc, EKRSACertificateHandle)
	if err != nil {
		err = fmt.Errorf("reading NV index %08x failed: %w", EKRSACertificateHandle, err)
		return "", err
	}

	ekCertData = bytes.Trim(ekCertData, "\xff")
	ekCertPEM := new(bytes.Buffer)

	err = pem.Encode(ekCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: ekCertData,
	})
	if err != nil {
		err = fmt.Errorf("reading NV index %08x failed: %w", EKRSACertificateHandle, err)
		return "", err
	}

	return ekCertPEM.String(), nil
}

func main() {
	rwc, err := tpm2.OpenTPM("/dev/tpmrm0")
	if err != nil {
		log.Fatalf("Errror opening TPM: %v", err)
	}

	defer func() {
		err = rwc.Close()
		if err != nil {
			log.Fatalf("Error closing TPM: %v", err)
		}
	}()

	ekPublicCertPEM, err := getEKPublicCertificate(rwc)
	if err != nil {
		log.Printf("Error retrieving EK Public Certificate: %v", err)
		return
	}

	fmt.Print(ekPublicCertPEM)
}
