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
)

// AuthorizeResponse provides the structure for the authorize response.
type AuthorizeResponse struct {
	Success bool   `json:"success"`
	Reason  string `json:"reason,omitempty"`
}

// Authorize handles the authorize api endpoint.
func Authorize(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	xname := r.FormValue("xname")
	nodeType := r.FormValue("type")

	var resp AuthorizeResponse

	log.Printf("Xname: %s  Type: %s", xname, nodeType)

	if err := validateXname(xname); err != nil {
		sendResponseError(w, err)
		return
	}

	sessionCookie, sessionExpiresAt := createSessionCookie(xname, nodeType)

	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   sessionCookie,
		Expires: sessionExpiresAt,
	})

	w.WriteHeader(http.StatusOK)

	resp = AuthorizeResponse{Success: true}

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("error encoding the authorize response: %v", err)
	}

	session := sessions[sessionCookie]
	session.step = 1
	sessions[sessionCookie] = session
}
