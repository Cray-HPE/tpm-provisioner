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
// Package provisioner provides the TPM Provisioner server functions.
package provisioner

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Route is an http route.
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes are a collection of http routes.
type Routes []Route

// NewRouter creates a new router.
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

var routes = Routes{
	{
		"Authorize",
		strings.ToUpper("Get"),
		"/apis/tpm-provisioner/authorize",
		Authorize,
	},
	{
		"RequestChallenge",
		strings.ToUpper("Post"),
		"/apis/tpm-provisioner/challenge/request",
		RequestChallenge,
	},
	{
		"SubmitChallenge",
		strings.ToUpper("Post"),
		"/apis/tpm-provisioner/challenge/submit",
		SubmitChallenge,
	},
	{
		"ClientPost",
		strings.ToUpper("Get"),
		"/apis/tpm-provisioner/whitelist/get",
		ListWhiteList,
	},
	{
		"ClientPost",
		strings.ToUpper("Post"),
		"/apis/tpm-provisioner/whitelist/add",
		AddWhiteList,
	},
	{
		"ClientPost",
		strings.ToUpper("Post"),
		"/apis/tpm-provisioner/whitelist/remove",
		RemoveWhiteList,
	},
}
