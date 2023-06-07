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
// Package client provides the TPM Providier client resources for use with the
// tpm-provider-client binary.
package client

import (
	"bytes"
	"context"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cray-hpe/tpm-provisioner/third_party/devid-provisioning-tool/pkg/agent/keygen"
	"github.com/cray-hpe/tpm-provisioner/third_party/devid-provisioning-tool/pkg/devid"
	tpm2tools "github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm/tpm2"
)

// getKeygen generates a new Keygen.
func getKeygen() *keygen.Keygen {
	srkTemplateHighRSA := tpm2tools.SRKTemplateRSA()
	srkTemplateHighRSA.RSAParameters.ModulusRaw = []byte{}

	return keygen.New(keygen.UseSRKTemplate(srkTemplateHighRSA))
}

// CreateRawRequest creates the raw challenge request.
func CreateRawRequest(ctx context.Context, rw io.ReadWriter, pi pkix.Name) (data,
	signature []byte, resources *devid.RequestResources, err error,
) {
	csr, resources, err := devid.CreateSigningRequest(ctx, getKeygen(), rw)
	if err != nil {
		err = fmt.Errorf("CSR creation failed: %w", err)
		return
	}

	defer func() {
		// Flush contexts on error
		if err != nil {
			resources.Flush()
			resources = nil
		}
	}()

	csr.PlatformIdentity = pi.ToRDNSequence()

	requestData, err := csr.MarshalBinary()
	if err != nil {
		err = fmt.Errorf("CSR marshal failed: %w", err)
		return
	}

	requestSig, err := devid.HashAndSign(rw, tpm2.HandleOwner, resources.DevID.Handle, requestData)
	if err != nil {
		err = fmt.Errorf("CSR signing failed: %w", err)
		return
	}

	return requestData, requestSig, resources, nil
}

// WriteDevID writes the devid certificate and the public and private blob files
// to the specificed directory.
func WriteDevID(outputDir string, resources *devid.RequestResources, devIDCert []byte) error {
	var devIDCertPem bytes.Buffer

	err := pem.Encode(&devIDCertPem, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: devIDCert,
	})
	if err != nil {
		return fmt.Errorf("certificate PEM encoding failed: %w", err)
	}

	dirPath := filepath.Dir(outputDir)

	_, err = os.Stat(dirPath)
	if os.IsNotExist(err) {
		err = os.Mkdir(dirPath, 0o700)
		if err != nil {
			return fmt.Errorf("unable to create output directory: %w", err)
		}
	}

	err = os.WriteFile(outputDir+"/devid.crt.pem", devIDCertPem.Bytes(), os.FileMode(0o600))
	if err != nil {
		return fmt.Errorf("writing DevID certificate at %q failed: %w", outputDir, err)
	}

	err = os.WriteFile(outputDir+"/devid.pub.blob", resources.DevID.PublicBlob, os.FileMode(0o600))
	if err != nil {
		return fmt.Errorf("writing DevID public key at %q failed: %w", outputDir, err)
	}

	err = os.WriteFile(outputDir+"/devid.priv.blob", resources.DevID.PrivateBlob, os.FileMode(0o600))
	if err != nil {
		return fmt.Errorf("writing DevID private key at %q failed: %w", outputDir, err)
	}

	return nil
}
