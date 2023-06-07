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
package verify

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"errors"
	"fmt"
	"log"

	"github.com/cray-hpe/tpm-provisioner/third_party/devid-provisioning-tool/pkg/devid"
	"github.com/google/go-tpm/tpm2"
)

type KeyAttributeError struct {
	Reason string
}

func (e KeyAttributeError) Error() string {
	return fmt.Sprintf("key attribute error: %s", e.Reason)
}

var subjectAlternativeNameOID = asn1.ObjectIdentifier{2, 5, 29, 17}

func ValidateRequest(data string, sig string, certPool *x509.CertPool) error {
	var sr devid.SigningRequest

	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}

	decodedSig, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		return err
	}

	err = sr.UnmarshalBinary(decodedData)
	if err != nil {
		return err
	}

	err = validateSignature(sr.DevIDKey, decodedData, decodedSig)
	if err != nil {
		return err
	}
	err = validateEndorcement(sr.EndorsementCertificate, certPool)
	if err != nil {
		return err
	}

	err = validateDevIDResidency(
		sr.AttestationKey,
		sr.DevIDKey,
		sr.CertifyData,
		sr.CertifySignature,
	)
	if err != nil {
		return err
	}

	err = checkDevIDProp(sr.DevIDKey.Attributes)
	if err != nil {
		return err
	}

	err = checkAKProp(sr.AttestationKey.Attributes)
	if err != nil {
		return err
	}

	return nil
}

// 7. CA verifies the received data:

// 7a. Extract IDevID public key and verify the signature on TCG-CSR-IDEVID
func validateSignature(pub *tpm2.Public, data []byte, sig []byte) error {
	key, err := pub.Key()
	if err != nil {
		return err
	}

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return errors.New("only RSA keys are supported")
	}

	sigScheme, err := devid.GetSignatureScheme(*pub)
	if err != nil {
		return err
	}

	hash, err := sigScheme.Hash.Hash()
	if err != nil {
		return err
	}

	h := hash.New()
	h.Write(data)
	hashed := h.Sum(nil)

	return rsa.VerifyPKCS1v15(rsaKey, hash, hashed, sig)
}

// 7b. Verify the EK Certificate using the indicated TPM manufacturer's
//
//	public key.
func validateEndorcement(cert *x509.Certificate, certPool *x509.CertPool) error {
	if len(cert.UnhandledCriticalExtensions) > 0 {
		unhandledExtensions := []asn1.ObjectIdentifier{}
		for _, oid := range cert.UnhandledCriticalExtensions {
			if oid.Equal(subjectAlternativeNameOID) {
				// Subject Alternative Name is not processed at the time.
				continue
			}
		}

		cert.UnhandledCriticalExtensions = unhandledExtensions
	}

	_, err := cert.Verify(x509.VerifyOptions{
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		Roots:     certPool,
	})
	if err != nil {
		return err
	}

	return nil
}

// 7c. Verify TPM residency of IDevID key using the IAK public key to
//
//	validate the signature of the TPMB_Attest structure.
func checkSignature(pub *tpm2.Public, data, sig []byte) error {
	key, err := pub.Key()
	if err != nil {
		return err
	}

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return errors.New("only RSA keys are supported")
	}

	sigScheme, err := devid.GetSignatureScheme(*pub)
	if err != nil {
		return err
	}

	hash, err := sigScheme.Hash.Hash()
	if err != nil {
		return err
	}

	h := hash.New()
	h.Write(data)
	hashed := h.Sum(nil)

	err = rsa.VerifyPKCS1v15(rsaKey, hash, hashed, sig)
	if err != nil {
		log.Printf("VerifypKCS1v15 Failed")
	}
	return nil
}

func validateDevIDResidency(AK *tpm2.Public, devIDPub *tpm2.Public, attestationData []byte, attestationSig []byte) error {
	err := checkSignature(AK, attestationData, attestationSig)
	if err != nil {
		return err
	}

	data, err := tpm2.DecodeAttestationData(attestationData)
	if err != nil {
		return err
	}

	if data.AttestedCertifyInfo == nil {
		return errors.New("missing certify info")
	}

	ok, err := data.AttestedCertifyInfo.Name.MatchesPublic(*devIDPub)
	if err != nil {
		return err
	}

	if !ok {
		return errors.New("certify failed")
	}

	return nil
}

// 7d. Verify the attributes of the IDevID key public area.
func checkDevIDProp(prop tpm2.KeyProp) error {
	if (prop & tpm2.FlagDecrypt) != 0 {
		return KeyAttributeError{
			Reason: "DevID should not be a decryption key",
		}
	}

	if (prop & tpm2.FlagRestricted) != 0 {
		return KeyAttributeError{
			Reason: "DevID should not be a restricted key",
		}
	}

	if (prop & tpm2.FlagSign) == 0 {
		return KeyAttributeError{
			Reason: "DevID should be a signing key",
		}
	}

	if (prop & tpm2.FlagFixedTPM) == 0 {
		return KeyAttributeError{
			Reason: "DevID should be fixedTPM",
		}
	}
	return nil
}

// 7e. Verify the attributes of the IAK public area.
func checkAKProp(prop tpm2.KeyProp) error {
	if (prop & tpm2.FlagDecrypt) != 0 {
		return KeyAttributeError{
			Reason: "AK should not be a decryption key",
		}
	}

	if (prop & tpm2.FlagRestricted) == 0 {
		return KeyAttributeError{
			Reason: "AK should be a restricted key",
		}
	}

	if (prop & tpm2.FlagSign) == 0 {
		return KeyAttributeError{
			Reason: "AK should be a signing key",
		}
	}

	if (prop & tpm2.FlagFixedTPM) == 0 {
		return KeyAttributeError{
			Reason: "AK should be fixedTPM",
		}
	}

	return nil
}
