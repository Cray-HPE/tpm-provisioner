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
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"

	"github.com/cray-hpe/tpm-provisioner/third_party/devid-provisioning-tool/pkg/devid"
	"github.com/google/go-tpm/legacy/tpm2/credactivation"
)

// CreateChallenge creates a challenge to be sent to the tpm-provisioner client.
func CreateChallenge(data string) (string, string, string, error) {
	var sr devid.SigningRequest

	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", "", "", err
	}

	err = sr.UnmarshalBinary(decodedData)
	if err != nil {
		return "", "", "", err
	}

	hash, err := sr.EndorsementKey.NameAlg.Hash()
	if err != nil {
		return "", "", "", err
	}

	credName, err := sr.AttestationKey.Name()
	if err != nil {
		return "", "", "", err
	}

	nonce := make([]byte, hash.Size())

	_, err = rand.Read(nonce)
	if err != nil {
		return "", "", "", err
	}

	encKey, err := sr.EndorsementKey.Key()
	if err != nil {
		return "", "", "", err
	}

	var symBlockSize int
	switch encKey.(type) {
	case *rsa.PublicKey:
		symBlockSize = int(sr.EndorsementKey.RSAParameters.Symmetric.KeyBits) / 8

	default:
		return "", "", "", errors.New("unsupported algorithm")
	}

	blob, secret, err := credactivation.Generate(
		credName.Digest,
		encKey,
		symBlockSize,
		nonce,
	)
	if err != nil {
		return "", "", "", err
	}

	blob64 := base64.RawStdEncoding.EncodeToString(blob[2:])
	secret64 := base64.RawStdEncoding.EncodeToString(secret[2:])
	nonce64 := base64.StdEncoding.EncodeToString(nonce)

	return blob64, secret64, nonce64, nil
}
