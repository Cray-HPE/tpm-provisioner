// Package main implements a simple way to retrieve the TPM's EK certificate.
package main

import (
	"bytes"
	"encoding/pem"
	"fmt"
	"io"
	"log"

	"github.com/google/go-tpm/legacy/tpm2"
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
