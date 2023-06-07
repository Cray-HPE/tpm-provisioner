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
package client

import (
	"encoding/base64"
	"io"

	"github.com/cray-hpe/tpm-provisioner/third_party/devid-provisioning-tool/pkg/devid"
	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

func createPolicySession(rw io.ReadWriter) (tpmutil.Handle, error) {
	var nonceCaller [32]byte

	hSession, _, err := tpm2.StartAuthSession(
		rw,
		tpm2.HandleNull,
		tpm2.HandleNull,
		nonceCaller[:],
		nil,
		tpm2.SessionPolicy,
		tpm2.AlgNull,
		tpm2.AlgSHA256,
	)
	if err != nil {
		return 0, err
	}

	_, _, err = tpm2.PolicySecret(
		rw,
		tpm2.HandleEndorsement,
		tpm2.AuthCommand{Session: tpm2.HandlePasswordSession},
		hSession,
		nil,
		nil,
		nil,
		0,
	)
	if err != nil {
		err = tpm2.FlushContext(rw, hSession)
		if err != nil {
			return 0, err
		}

		return 0, err
	}

	return hSession, nil
}

// GenerateChallengeResponse creates a challenge response to send to the TPM
// Provisioner.
func GenerateChallengeResponse(rw io.ReadWriter, blob string, secret string, resources *devid.RequestResources) ([]byte, error) {
	decodedBlob, err := base64.RawStdEncoding.DecodeString(blob)
	if err != nil {
		return nil, err
	}

	decodedSecret, err := base64.RawStdEncoding.DecodeString(secret)
	if err != nil {
		return nil, err
	}

	session, err := createPolicySession(rw)
	if err != nil {
		return nil, err
	}

	return tpm2.ActivateCredentialUsingAuth(
		rw,
		[]tpm2.AuthCommand{
			{Session: tpm2.HandlePasswordSession},
			{Session: session},
		},
		resources.Attestation.Handle,
		resources.Endorsement.Handle,
		decodedBlob,
		decodedSecret,
	)
}
