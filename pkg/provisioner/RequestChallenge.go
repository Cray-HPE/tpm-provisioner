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
	"encoding/json"
	"log"
	"net/http"

	"github.com/cray-hpe/tpm-provisioner/pkg/verify"
)

// CertificateResponse contains the response structure for the certificate
// request.
type CertificateResponse struct {
	Success bool   `json:"success"`
	Reason  string `json:"reason,omitempty"`
	Blob    string `json:"blob"`
	Secret  string `json:"secret"`
}

// CertificateRequest contains the certificate request structure.
type CertificateRequest struct {
	Data string `json:"data"`
	Sig  string `json:"sig"`
}

// RequestChallenge handles the challenge request api.
func RequestChallenge(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	err := validateCookie(r.Cookies(), 1)
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

	err = verify.ValidateRequest(data.Data, data.Sig, CFG.ManufactuerCAs)
	if err != nil {
		sendResponseError(w, err)
		return
	}

	blob, secret, nonce, err := CreateChallenge(data.Data)
	if err != nil {
		sendResponseError(w, err)
		return
	}

	err = setNonce(r.Cookies(), nonce)

	if err != nil {
		sendResponseError(w, err)
		return
	}

	err = setReqData(r.Cookies(), data.Data)
	if err != nil {
		sendResponseError(w, err)
		return
	}

	certResp := CertificateResponse{
		Success: true,
		Blob:    blob,
		Secret:  secret,
	}

	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(certResp)
	if err != nil {
		log.Printf("error encoding the certificate response: %v", err)
	}
}
