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
	"log"
	"os"
	"path/filepath"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// Store stores the DevID certificate and blobs into a TPM.
func Store(rwc io.ReadWriter, cfg Config) error {
	err := storeFile(rwc, cfg.DevCertPath, cfg.DevCertAddr)
	if err != nil {
		return err
	}

	err = storeFile(rwc, cfg.DevPubBlobPath, cfg.DevPubBlobAddr)
	if err != nil {
		return err
	}

	err = storeFile(rwc, cfg.DevPrivBlobPath, cfg.DevPrivBlobAddr)
	if err != nil {
		return err
	}

	return nil
}

func storeFile(rwc io.ReadWriter, file string, addr int) error {
	var data []byte

	fmt.Printf("Writing %s to NVRam (%x)\n", file, addr)

	data, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		return err
	}

	err = SaveData(rwc, addr, data)
	if err != nil {
		return err
	}

	return nil
}

// SaveData saves a []byte to the TPM at the specified address.
func SaveData(rwc io.ReadWriter, addr int, data []byte) error {
	attr := tpm2.AttrOwnerWrite | tpm2.AttrOwnerRead | tpm2.AttrWriteSTClear | tpm2.AttrReadSTClear
	idx := tpmutil.Handle(addr)

	err := tpm2.NVDefineSpace(rwc, tpm2.HandleOwner, idx, "", "", nil, attr, uint16(len(data)))
	if err != nil {
		return err
	}

	for i := 0; i < len(data); i += 1024 {
		var maxLen int
		if len(data)-i >= 1024 {
			maxLen = i + 1024
		} else {
			maxLen = len(data)
		}

		fmt.Printf("Saving Data of size: %d\n", maxLen)

		err = tpm2.NVWrite(rwc, tpm2.HandleOwner, idx, "", data[i:maxLen], uint16(i))
		if err != nil {
			log.Printf("NVWrite Error: %v", err)
			return err
		}
	}

	return nil
}
