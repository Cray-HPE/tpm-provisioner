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
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/cray-hpe/tpm-provisioner/third_party/devid-provisioning-tool/pkg/common"
	"github.com/cray-hpe/tpm-provisioner/third_party/devid-provisioning-tool/pkg/devid"
	"github.com/cray-hpe/tpm-provisioner/third_party/devid-provisioning-tool/pkg/x509tcg"
)

// SubmitResponse contains the response to the submit challenge api request.
type SubmitResponse struct {
	Success          bool   `json:"success"`
	Reason           string `json:"reason,omitempty"`
	DevIDCertificate string `json:"devIdCertificate"`
}

// SubmitRequest contains the request for the submit challenge api.
type SubmitRequest struct {
	Data string `json:"data"`
}

// SubmitChallenge handles the challenge/submit api request.
func SubmitChallenge(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var submitResp SubmitResponse

	err := validateCookie(r.Cookies(), 2)
	if err != nil {
		sendResponseError(w, err)
		return
	}

	decoder := json.NewDecoder(r.Body)

	var data CertificateRequest

	err = decoder.Decode(&data)
	if err != nil {
		sendResponseError(w, err)
		return
	}

	xname, err := getXname(r.Cookies())
	if err != nil {
		sendResponseError(w, err)
		return
	}

	nodeType, err := getType(r.Cookies())
	if err != nil {
		sendResponseError(w, err)
		return
	}

	nonce, err := getNonce(r.Cookies())
	if err != nil {
		sendResponseError(w, err)
		return
	}

	if data.Data != nonce {
		sendResponseError(w, errors.New("challenge response does not match nonce"))
		return
	}

	reqData, err := getReqData(r.Cookies())
	if err != nil {
		sendResponseError(w, err)
		return
	}

	decodedReqData, err := base64.StdEncoding.DecodeString(reqData)
	if err != nil {
		sendResponseError(w, err)
		return
	}

	devIDCert, err := issueDevIDCertificate(decodedReqData)
	if err != nil {
		sendResponseError(w, err)
		return
	}

	submitResp = SubmitResponse{
		Success:          true,
		DevIDCertificate: devIDCert,
	}

	w.WriteHeader(http.StatusInternalServerError)

	err = json.NewEncoder(w).Encode(submitResp)
	if err != nil {
		log.Printf("error encoding the response: %v", err)
		return
	}

	err = requestSpireWorkloads(nodeType, xname, CFG.SpireTokensURL)
	if err != nil {
		log.Printf("error requesting the creation of spire workloads: %v", err)
		return
	}
}

func issueDevIDCertificate(data []byte) (string, error) {
	var subExtras *common.DistinguishedName

	var sr devid.SigningRequest

	err := sr.UnmarshalBinary(data)
	if err != nil {
		return "", err
	}

	if sr.DevIDKey == nil {
		return "", errors.New("missing DevID key")
	}

	pub, err := sr.DevIDKey.Key()
	if err != nil {
		return "", err
	}

	keyData, err := sr.DevIDKey.Encode()
	if err != nil {
		return "", err
	}

	var subj pkix.Name

	subj.FillFromRDNSequence(&sr.PlatformIdentity)

	subExtras.AppendInto(&subj)

	subjectIsEmpty := len(subj.ToRDNSequence()) == 0

	sanExtension, err := x509tcg.DevIDSANFromEKCertificate(
		subjectIsEmpty,
		sr.EndorsementCertificate,
	)
	if err != nil {
		return "", err
	}

	keySha256 := sha256.Sum256(keyData)
	serialNumber := new(big.Int).SetBytes(keySha256[:])

	template := x509.Certificate{
		SerialNumber: serialNumber,
		PublicKey:    pub,

		Subject:   subj,
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 0, 0),

		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  false,

		UnknownExtKeyUsage: []asn1.ObjectIdentifier{{2, 23, 133, 11, 1, 2}},

		ExtraExtensions: []pkix.Extension{
			sanExtension,
		},
	}

	cert, err := x509.CreateCertificate(rand.Reader, &template, CFG.ProviderCA, template.PublicKey, CFG.ProviderKey)
	if err != nil {
		return "", err
	}

	encodedCert := base64.RawStdEncoding.EncodeToString(cert)

	return encodedCert, nil
}
