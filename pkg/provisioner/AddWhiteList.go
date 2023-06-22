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

// AddWhiteListResponse  contains the response structure for the add whitelist
// api request.
type AddWhiteListResponse struct {
	Success bool   `json:"success"`
	Reason  string `json:"reason,omitempty"`
}

// AddWhiteList handles the add xname regexp to white list api endpoint.
func AddWhiteList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	str := r.FormValue("xname")

	var resp AddWhiteListResponse

	err := AddWhiteListItem(CFG.WhiteList, str)
	if err != nil {
		resp = AddWhiteListResponse{
			Success: false,
			Reason:  err.Error(),
		}

		w.WriteHeader(http.StatusInternalServerError)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Printf("Error encoding json response: %v", err)
		}

		return
	}

	resp = AddWhiteListResponse{
		Success: true,
	}

	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("Error encoding json response: %v", err)
	}
}
