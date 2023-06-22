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
// tpm-provisioner-server signs DevIDs for TPMs for use with spire.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cray-hpe/tpm-provisioner/pkg/provisioner"
	"github.com/gorilla/mux"
)

func requestLoggerMiddleware(r *mux.Router) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			defer func() {
				log.Printf(
					"[%s] [%v] %s %s %s",
					req.Method,
					time.Since(start),
					req.Host,
					req.URL.Path,
					req.URL.RawQuery,
				)
			}()
			next.ServeHTTP(w, req)
		})
	}
}

func main() {
	log.Printf("Server started")

	if len(os.Args) > 2 {
		log.Fatalf("%s [CONFIG FILE]", os.Args[0])
	}

	var f string

	var p string

	if len(os.Args) == 1 {
		p = "/etc/tpm-provisioner"
		f = "server.conf"
	} else {
		s := strings.Split(os.Args[1], "/")
		p = strings.Join(s[:len(s)-1], "/")
		f = s[len(s)-1]
	}

	err := provisioner.ParseConfig(p, f)
	if err != nil {
		log.Fatalf("Unable to parse config: %v", err)
	}

	err = provisioner.LoadWhiteList(provisioner.CFG.WhiteList)
	if err != nil {
		log.Fatalf("Unable to load whitelist: %v", err)
	}

	router := provisioner.NewRouter()
	address := fmt.Sprintf(":%d", provisioner.CFG.Port)

	router.Use(requestLoggerMiddleware(router))

	srv := http.Server{
		Addr:              address,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	go provisioner.CleanSessions()

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
