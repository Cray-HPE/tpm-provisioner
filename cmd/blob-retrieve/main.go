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
// blob-retrieve retrieves a set of files that is stored in a TPM's NVRam.
// This uses a configuration file that specifies the set of blobs and their
// location in the TPM.
// By default it uses the /etc/tpm-provisioner/blobs.conf config file
package main

import (
	"io"
	"log"
	"os"
	"strings"

	"github.com/cray-hpe/tpm-provisioner/pkg/blob"
	"github.com/google/go-tpm/tpm2"
)

// openTPM opens the tpm for use.
func openTPM() io.ReadWriter {
	rwc, err := tpm2.OpenTPM("/dev/tpmrm0")
	if err != nil {
		log.Fatalf("Error opening TPM: %v", err)
	}

	return rwc
}

func main() {
	if len(os.Args) > 2 {
		log.Fatalf("%s [CONFIG FILE]", os.Args[0])
	}

	rwc := openTPM()

	var f string

	var p string

	if len(os.Args) == 1 {
		p = "/etc/tpm-provisioner"
		f = "blobs.conf"
	} else {
		s := strings.Split(os.Args[1], "/")
		p = strings.Join(s[:len(s)-1], "/")
		f = s[len(s)-1]
	}

	cfg, err := blob.ParseConfig(p, f)
	if err != nil {
		log.Fatalf("Unable to parse config: %v", err)
	}

	err = blob.Retrieve(rwc, cfg)
	if err != nil {
		log.Fatalf("Unable to retrieve blobs: %v", err)
	}
}
