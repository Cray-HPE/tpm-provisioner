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
// Package blob stores and retrieves data from a TPM's NVRam.
package blob

import (
	"fmt"
	"io"
	"os"

	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// Retrieve reads the DevID certificate and blobs from a TPM and stores them in
// the files specified in the config.
func Retrieve(rwc io.ReadWriter, cfg Config) error {
	err := retrieveFile(rwc, cfg.DevCertPath, cfg.DevCertAddr)
	if err != nil {
		return err
	}

	err = retrieveFile(rwc, cfg.DevPubBlobPath, cfg.DevPubBlobAddr)
	if err != nil {
		return err
	}

	err = retrieveFile(rwc, cfg.DevPrivBlobPath, cfg.DevPrivBlobAddr)
	if err != nil {
		return err
	}

	return nil
}

func retrieveFile(rwc io.ReadWriter, path string, addr int) error {
	fmt.Printf("Retrieving %s from %x\n", path, addr)

	var data []byte

	data, err := LoadData(rwc, addr)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, data, 0o600)
	if err != nil {
		return err
	}

	return nil
}

// LoadData returns []byte read from the addr in a TPM.
func LoadData(rwc io.ReadWriter, addr int) ([]byte, error) {
	nvData, err := tpm2.NVReadEx(rwc, tpmutil.Handle(addr), tpm2.HandleOwner, "", 0)
	if err != nil {
		err = fmt.Errorf("tpm2.ReadPublic failed: %w", err)
		return nil, err
	}

	fmt.Printf("Retrieved Data of size: %d\n", len(nvData))

	return nvData, nil
}
